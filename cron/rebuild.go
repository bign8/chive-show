package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/bign8/chive-show/appengine"
	"github.com/bign8/chive-show/models"
	"github.com/bign8/chive-show/models/datastore"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

func RebuildHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bits, err := httputil.DumpRequest(r, true)
		if err != nil {
			panic(err)
		}
		w.Write(bits)
		// http.Error(w, "TODO", http.StatusNotImplemented)
	}
}

const (
	// SIZE of a batch
	SIZE = 2

	// DEBUG enable if troubleshooting algorithm
	DEBUG = true

	// DEPTH depth of feed mining
	DEPTH = 1

	// DEFERRED if deferreds should be processed deferred
	DEFERRED = true
)

func parse(store *datastore.Store, tasker *cloudtasks.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fp := new(feedParser)
		err := fp.Main(r.Context(), store, tasker, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			fmt.Fprint(w, "Parsed")
		}
	}
}

type feedParser struct {
	context context.Context
	store   *datastore.Store
	tasker  *cloudtasks.Client

	todo  []int
	posts []models.Post
}

func (x *feedParser) Main(c context.Context, store *datastore.Store, tasker *cloudtasks.Client, w http.ResponseWriter) error {
	x.context = c
	x.store = store
	x.tasker = tasker
	x.posts = make([]models.Post, 0)

	// Initial recursive edge case
	isStop, fullStop, err := x.isStop(1)
	if isStop || fullStop || err != nil {
		log.Printf("INFO: Finished without recursive searching %v", err)
		if err == nil {
			err = x.store.PutMulti(x.context, x.posts)
		}
		return err
	}

	// Recursive search strategy
	err = x.Search(1, -1)

	// storePosts and processTodo
	if err == nil {
		errc := make(chan error)
		go func() {
			errc <- x.store.PutMulti(x.context, x.posts)
		}()
		go func() {
			errc <- x.processTodo()
		}()
		err1, err2 := <-errc, <-errc
		if err1 != nil {
			err = err1
		} else if err2 != nil {
			err = err2
		}
	}

	if err != nil {
		log.Printf("Error: in Main %v", err)
	}
	return err
}

func batch(store *datastore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: ensure this task is coming from appengine
		if r.Method != http.MethodPost {
			log.Printf("Batch: got a %s request", r.Method)
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		var ids []int
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			log.Printf("Batch: unmarshal error: %v", err)
			http.Error(w, "invalid payload", http.StatusExpectationFailed)
			return
		}

		parser := feedParser{
			context: r.Context(),
			store:   store,
		}
		parser.processBatch(ids)
	}
}

func (x *feedParser) enqueueBatch(ids []int) error {
	body, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#AppEngineHttpRequest
	_, err = x.tasker.CreateTask(x.context, &taskspb.CreateTaskRequest{
		Parent: "projects/" + appengine.ProjectID() + "/locations/us-central1/queues/default",
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_POST,
					RelativeUri: "/cron/batch",
					Body:        body, // 100kb limit
				},
			},
		},
	})
	return err
}

func (x *feedParser) processBatch(ids []int) error {
	done := make(chan error)
	for _, idx := range ids {
		go func(idx int) {
			_, posts, err := getAndParseFeed(x.context, x.store, idx)
			if err == nil {
				err = x.store.PutMulti(x.context, posts)
			}
			done <- err
		}(idx)
	}
	for i := 0; i < len(ids); i++ {
		err := <-done
		if err != nil {
			log.Printf("error storing feed (at index %d): %v", i, err)
			return err
		}
	}
	return nil
}

func (x *feedParser) processTodo() error {
	log.Printf("INFO: Processing TODO: %v", x.todo)
	// TODO: use slice offsets into x.todo array rather than creating batch arrays

	var batch []int
	var err error
	for _, idx := range x.todo {
		if batch == nil {
			batch = make([]int, 0)
		}
		batch = append(batch, idx)
		if len(batch) >= SIZE {
			if DEFERRED {
				err = x.enqueueBatch(batch)
			} else {
				err = x.processBatch(batch)
			}
			if err != nil {
				return err
			}
			batch = nil
		}
	}
	if len(batch) > 0 {
		if DEFERRED {
			err = x.enqueueBatch(batch)
		} else {
			err = x.processBatch(batch)
		}
	}
	return err
}

func (x *feedParser) addRange(bottom, top int) {
	for i := bottom + 1; i < top; i++ {
		x.todo = append(x.todo, i)
	}
}

func (x *feedParser) Search(bottom, top int) (err error) {
	/*
	  def infinite_length(bottom=1, top=-1):
	    if bottom == 1 and not item_exists(1): return 0  # Starting edge case
	    if bottom == top - 1: return bottom  # Result found! (top doesn’t exist)
	    if top < 0:  # Searching forward
	      top = bottom << 1  # Base 2 hops
	      if item_exists(top):
	        top, bottom = -1, top # continue searching forward
	    else:  # Binary search between bottom and top
	      middle = (bottom + top) // 2
	      bottom, top = middle, top if item_exists(middle) else bottom, middle
	    return infinite_length(bottom, top)  # Tail recursion!!!
	*/
	if bottom == top-1 {
		log.Printf("INFO: TOP OF RANGE FOUND! @%d", top)
		x.addRange(bottom, top)
		return nil
	}
	var fullStop, isStop bool = false, false
	if top < 0 { // Searching forward
		top = bottom << 1 // Base 2 hops forward
		isStop, fullStop, err = x.isStop(top)
		if err != nil {
			return err
		}
		if !isStop {
			x.addRange(bottom, top)
			top, bottom = -1, top
		}
	} else { // Binary search between top and bottom
		middle := (bottom + top) / 2
		isStop, fullStop, err = x.isStop(middle)
		if err != nil {
			return err
		}
		if isStop {
			top = middle
		} else {
			x.addRange(bottom, middle)
			bottom = middle
		}
	}
	if fullStop {
		return nil
	}
	return x.Search(bottom, top) // TAIL RECURSION!!!
}

func (x *feedParser) isStop(idx int) (isStop, fullStop bool, err error) {
	// Gather posts as necessary
	found, posts, err := getAndParseFeed(x.context, x.store, idx)
	if err == ErrFeedParse404 {
		log.Printf("INFO: Reached the end of the feed list (%v)", idx)
		return true, false, nil
	}
	if err != nil {
		log.Printf("Error decoding ChiveFeed: %s", err)
		return false, false, err
	}
	x.posts = append(x.posts, posts...)

	// Use store_count info to determine if isStop
	count := len(posts)
	isStop = count == 0 || DEBUG
	fullStop = found != count && count > 0
	if DEBUG {
		isStop = idx > DEPTH
		fullStop = idx == DEPTH
	}
	return
}
