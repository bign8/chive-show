package cron

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/datastore"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/anaskhan96/soup"
	"go.opencensus.io/plugin/ochttp"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/bign8/chive-show/keycache"
	"github.com/bign8/chive-show/models"
)

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

var client = &http.Client{
	Transport: &ochttp.Transport{
		// Use Google Cloud propagation format.
		Propagation: &propagation.HTTPFormat{},
	},
}

// Init initializes cron handlers
func Init(store *datastore.Client) {
	tasker, err := cloudtasks.NewClient(context.Background())
	if err != nil {
		panic(err)
	}
	http.Handle("/cron/parse", parse(store, tasker))
	http.Handle("/cron/batch", batch(store))
	http.Handle("/cron/delete", delete(store))
}

var (
	// ErrFeedParse404 if feed page is not found
	ErrFeedParse404 = fmt.Errorf("Feed parcing recieved a %d Status Code", 404)
)

func pageURL(idx int) string {
	return fmt.Sprintf("http://thechive.com/feed/?paged=%d", idx)
}

func parse(store *datastore.Client, tasker *cloudtasks.Client) http.HandlerFunc {
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
	store   *datastore.Client
	tasker  *cloudtasks.Client

	todo  []int
	guids map[int64]bool // this could be extremely large
	posts []models.Post
}

func (x *feedParser) Main(c context.Context, store *datastore.Client, tasker *cloudtasks.Client, w http.ResponseWriter) error {
	x.context = c
	x.store = store
	x.tasker = tasker

	// Load guids from DB
	// TODO: do this with sharded keys
	keys, err := store.GetAll(c, datastore.NewQuery(models.POST).KeysOnly(), nil)
	if err != nil {
		log.Printf("Error: finding keys %v", err)
		return err
	}
	x.guids = map[int64]bool{}
	for _, key := range keys {
		x.guids[key.ID] = true
	}
	keys = nil

	// // DEBUG ONLY
	// data, err := json.MarshalIndent(x.guids, "", "  ")
	// fmt.Fprint(w, string(data))
	// return err
	x.posts = make([]models.Post, 0)

	// Initial recursive edge case
	isStop, fullStop, err := x.isStop(1)
	if isStop || fullStop || err != nil {
		log.Printf("INFO: Finished without recursive searching %v", err)
		if err == nil {
			err = x.storePosts(x.posts)
		}
		return err
	}

	// Recursive search strategy
	err = x.Search(1, -1)

	// storePosts and processTodo
	if err == nil {
		errc := make(chan error)
		go func() {
			errc <- x.storePosts(x.posts)
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

func batch(store *datastore.Client) http.HandlerFunc {
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
		Parent: "projects/crucial-alpha-706/locations/us-central1/queues/default",
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_POST,
					RelativeUri: "/cron/batch",
					Body:        body,
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
			posts, err := x.getAndParseFeed(idx)
			if err == nil {
				err = x.storePosts(posts)
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
	posts, err := x.getAndParseFeed(idx)
	if err == ErrFeedParse404 {
		log.Printf("INFO: Reached the end of the feed list (%v)", idx)
		return true, false, nil
	}
	if err != nil {
		log.Printf("Error decoding ChiveFeed: %s", err)
		return false, false, err
	}

	// Check for Duplicates
	count := 0
	for _, post := range posts {
		id, _, err := guidToInt(post.GUID)
		if x.guids[id] || err != nil {
			continue
		}
		count++
	}
	x.posts = append(x.posts, posts...)

	// Use store_count info to determine if isStop
	isStop = count == 0 || DEBUG
	fullStop = len(posts) != count && count > 0
	if DEBUG {
		isStop = idx > DEPTH
		fullStop = idx == DEPTH
	}
	return
}

func (x *feedParser) getAndParseFeed(idx int) ([]models.Post, error) {
	url := pageURL(idx)

	// Get Response
	log.Printf("INFO: Parsing index %v (%v)", idx, url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req.WithContext(x.context))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, ErrFeedParse404
		}
		return nil, fmt.Errorf("feed parcing recieved a %d Status Code", resp.StatusCode)
	}

	// Decode Response
	decoder := xml.NewDecoder(resp.Body)
	var feed struct {
		Items []models.Post `xml:"channel>item"`
	}
	if decoder.Decode(&feed) != nil {
		return nil, err
	}

	// Cleanup Response
	for idx := range feed.Items {
		post := &feed.Items[idx]
		if err = mine(post); err != nil {
			log.Printf("Unable to mine page for details: %s", post.Link)
		}
		for i, img := range post.Media {
			post.Media[i].URL = stripQuery(img.URL)
		}

		// Find the "author" image and remove it from the "media" set
		var found bool
		for i, media := range post.Media {
			if media.Category == "author" {
				found = true
				post.MugShot = media.URL
				post.Media = append(post.Media[:i], post.Media[i+1:]...)
				break
			}
		}
		if !found {
			log.Printf("Unable to find author for: %#v", post)
			post.MugShot = post.Media[0].URL
			post.Media = post.Media[1:]
		}
	}
	return feed.Items, err
}

func mine(post *models.Post) error {
	res, err := soup.Get(post.Link)
	if err != nil {
		log.Printf("mine(%s): unable to fetch: %s", post.Link, err)
		return nil
	}
	doc := soup.HTMLParse(res)

	for _, figure := range doc.FindAll("figure") {
		obj := figure.Find("img")
		if obj.Error != nil {
			log.Printf("Unable to find img: %s: %v", post.Link, obj.Error)
			continue
		}
		post.Media = append(post.Media, models.Img{
			URL: obj.Attrs()["src"],
		})
	}

	// parse CHIVE_GALLERY_ITEMS from script id='chive-theme-js-js-extra' into JSON
	// TODO: use match the image prefix? "https:\/\/thechive.com\/wp-content\/uploads\/" in the HTML and parse to closing "
	src := doc.Find("script", "id", "chive-theme-js-js-extra").FullText()
	idx := strings.IndexByte(src, '{')
	if idx < 0 {
		return errors.New("unable to find opening brace")
	}
	src = src[idx:]
	idx = strings.LastIndexByte(src, '}')
	if idx < 0 {
		return errors.New("unable to find closing brace")
	}
	src = src[:idx+1]
	var what struct {
		Items []struct {
			HTML *string `json:"html,omitempty"`
		} `json:"items"`
	}
	err = json.Unmarshal([]byte(src), &what)
	if err != nil {
		panic(err)
	}

	// Parse HTML attributes of JSON to get images
	for i, obj := range what.Items {
		if obj.HTML == nil {
			log.Printf("no HTML found in fragment %d (embedded in post?)", i)
			continue
		}
		ele := soup.HTMLParse(*obj.HTML)
		if ele.Error != nil {
			log.Printf("unable to parse post fragment %d", i)
			continue
		}
		imgs := ele.FindAll("img")
		if len(imgs) == 0 {
			log.Printf("no images found in fragment %d (video?)", i)
			continue
		}
		for _, img := range imgs {
			post.Media = append(post.Media, models.Img{
				URL: img.Attrs()["src"],
			})
		}
	}
	return nil
}

func (x *feedParser) storePosts(dirty []models.Post) error {
	var posts []models.Post
	var keys []*datastore.Key
	for _, post := range dirty {
		key, err := x.cleanPost(&post)
		if err != nil {
			log.Printf("Unable to clean post: %v %v", post.GUID, err)
			continue
		}
		posts = append(posts, post)
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}
	complete, err := x.store.PutMulti(x.context, keys, posts)
	if err != nil {
		log.Printf("Problem storing Post: %v %v", keys, err)
		return err
	}
	return keycache.AddKeys(x.context, x.store, models.POST, complete)
}

func (x *feedParser) cleanPost(p *models.Post) (*datastore.Key, error) {
	id, link, err := guidToInt(p.GUID)
	if err != nil {
		return nil, err
	}
	// Remove link posts
	if link {
		log.Printf("INFO: Ignoring links post %v \"%v\"", p.GUID, p.Title)
		return nil, fmt.Errorf("Ignoring links post")
	}

	// Detect video only posts
	video := regexp.MustCompile("\\([^&]*Video.*\\)")
	if video.MatchString(p.Title) {
		log.Printf("INFO: Ignoring video post %v \"%v\"", p.GUID, p.Title)
		return nil, fmt.Errorf("Ignoring video post")
	}
	log.Printf("INFO: Storing post %v \"%v\"", p.GUID, p.Title)

	// Cleanup post titles
	clean := regexp.MustCompile("\\W\\(([^\\)]*)\\)$")
	p.Title = clean.ReplaceAllLiteralString(p.Title, "")

	// Post
	// temp_key := datastore.NewIncompleteKey(x.context, DB_POST_TABLE, nil)
	key := datastore.IDKey(models.POST, id, nil)
	return key, nil
}

func guidToInt(guid string) (int64, bool, error) {
	// Remove link posts
	url, err := url.Parse(guid)
	if err != nil {
		return -1, false, err
	}

	// Parsing post id from guid url
	id, err := strconv.Atoi(url.Query().Get("p"))
	if err != nil {
		return -1, false, err
	}
	return int64(id), url.Query().Get("post_type") == "sdac_links", nil
}

func stripQuery(dirty string) string {
	obj, err := url.Parse(dirty)
	if err != nil {
		return dirty
	}
	obj.RawQuery = ""
	return obj.String()
}
