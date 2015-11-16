package cron

import (
	"fmt"
	"net/http"

	"github.com/bign8/chive-show/app/cron/crawler"
	"github.com/bign8/chive-show/app/cron/proj"

	"gopkg.in/mjibson/v1/appstats"

	"appengine"
	"appengine/datastore"
)

// const (
// 	// SIZE of a batch
// 	SIZE = 10
//
// 	// DEBUG enable if troubleshooting algorithm
// 	DEBUG = true
//
// 	// DEPTH depth of feed mining
// 	DEPTH = 1
//
// 	// DEFERRED if deferreds should be processed deferred
// 	DEFERRED = true
// )

func cleanup(c appengine.Context, name string) error {
	c.Infof("Cleaning %s", name)
	q := datastore.NewQuery(name).KeysOnly()
	keys, err := q.GetAll(c, nil)
	s := 100
	for len(keys) > 0 {
		if len(keys) < 100 {
			s = len(keys)
		}
		err = datastore.DeleteMulti(c, keys[:s])
		keys = keys[s:]
	}
	return err
}

// Init initializes cron handlers
func Init() {
	http.HandleFunc("/cron/stage/1", crawler.Crawl)

	http.Handle("/proj/tags", appstats.NewHandler(proj.Tags))
	http.Handle("/proj/graph", appstats.NewHandler(proj.Graph))
	http.Handle("/proj/shard", appstats.NewHandler(proj.TestShard))

	http.HandleFunc("/clean", func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		cleanup(c, "buff")
		cleanup(c, "edge")
		cleanup(c, "vertex")
		cleanup(c, "post")
	})

	http.Handle("/cron/stats", appstats.NewHandler(crawler.Stats))

	// http.Handle("/cron/parse", appstats.NewHandler(parseFeeds))
	// http.HandleFunc("/cron/delete", delete)
	http.HandleFunc("/_ah/start", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Start")
	})
	http.HandleFunc("/_ah/stop", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Stop")
	})
}

// var (
// 	// ErrFeedParse404 if feed page is not found
// 	ErrFeedParse404 = fmt.Errorf("Feed parcing recieved a %d Status Code", 404)
// )
//
// func pageURL(idx int) string {
// 	return fmt.Sprintf("http://thechive.com/feed/?paged=%d", idx)
// }
//
// func parseFeeds(c appengine.Context, w http.ResponseWriter, r *http.Request) {
// 	fp := new(feedParser)
// 	err := fp.Main(c, w)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	} else {
// 		fmt.Fprint(w, "Parsed")
// 	}
// }
//
// type feedParser struct {
// 	context appengine.Context
// 	client  *http.Client
//
// 	todo  []int
// 	guids map[int64]bool // this could be extremely large
// 	posts []models.Post
// }
//
// func (x *feedParser) Main(c appengine.Context, w http.ResponseWriter) error {
// 	x.context = c
// 	x.client = urlfetch.Client(c)
//
// 	// Load guids from DB
// 	// TODO: do this with sharded keys
// 	keys, err := datastore.NewQuery(models.POST).KeysOnly().GetAll(c, nil)
// 	if err != nil {
// 		c.Errorf("Error finding keys %v %v", err, appengine.IsOverQuota(err))
// 		return err
// 	}
// 	x.guids = map[int64]bool{}
// 	for _, key := range keys {
// 		x.guids[key.IntID()] = true
// 	}
// 	keys = nil
//
// 	// // DEBUG ONLY
// 	// data, err := json.MarshalIndent(x.guids, "", "  ")
// 	// fmt.Fprint(w, string(data))
// 	// return err
// 	x.posts = make([]models.Post, 0)
//
// 	// Initial recursive edge case
// 	isStop, fullStop, err := x.isStop(1)
// 	if isStop || fullStop || err != nil {
// 		c.Infof("Finished without recursive searching %v", err)
// 		if err == nil {
// 			err = x.storePosts(x.posts)
// 		}
// 		return err
// 	}
//
// 	// Recursive search strategy
// 	err = x.Search(1, -1)
//
// 	// storePosts and processTodo
// 	if err == nil {
// 		errc := make(chan error)
// 		go func() {
// 			errc <- x.storePosts(x.posts)
// 		}()
// 		go func() {
// 			errc <- x.processTodo()
// 		}()
// 		err1, err2 := <-errc, <-errc
// 		if err1 != nil {
// 			err = err1
// 		} else if err2 != nil {
// 			err = err2
// 		}
// 	}
//
// 	if err != nil {
// 		c.Errorf("Error in Main %v", err)
// 	}
// 	return err
// }
//
// var processBatchDeferred = delay.Func("process-todo-batch", func(c appengine.Context, ids []int) {
// 	parser := feedParser{
// 		context: c,
// 		client:  urlfetch.Client(c),
// 	}
// 	parser.processBatch(ids)
// })
//
// func (x *feedParser) processBatch(ids []int) error {
// 	done := make(chan error)
// 	for _, idx := range ids {
// 		go func(idx int) {
// 			posts, err := x.getAndParseFeed(idx)
// 			if err == nil {
// 				err = x.storePosts(posts)
// 			}
// 			done <- err
// 		}(idx)
// 	}
// 	for i := 0; i < len(ids); i++ {
// 		err := <-done
// 		if err != nil {
// 			x.context.Errorf("error storing feed (at index %d): %v", i, err)
// 			return err
// 		}
// 	}
// 	return nil
// }
//
// func (x *feedParser) processTodo() error {
// 	x.context.Infof("Processing TODO: %v", x.todo)
//
// 	var batch []int
// 	var task *taskqueue.Task
// 	var allTasks []*taskqueue.Task
// 	var err error
// 	for _, idx := range x.todo {
// 		if batch == nil {
// 			batch = make([]int, 0)
// 		}
// 		batch = append(batch, idx)
// 		if len(batch) >= SIZE {
// 			if DEFERRED {
// 				task, err = processBatchDeferred.Task(batch)
// 				if err == nil {
// 					allTasks = append(allTasks, task)
// 				}
// 			} else {
// 				err = x.processBatch(batch)
// 			}
// 			if err != nil {
// 				return err
// 			}
// 			batch = nil
// 		}
// 	}
// 	if len(batch) > 0 {
// 		if DEFERRED {
// 			task, err = processBatchDeferred.Task(batch)
// 			if err == nil {
// 				allTasks = append(allTasks, task)
// 			}
// 		} else {
// 			err = x.processBatch(batch)
// 		}
// 	}
// 	if DEFERRED && len(allTasks) > 0 {
// 		x.context.Infof("Adding %d task(s) to the default queue", len(allTasks))
// 		taskqueue.AddMulti(x.context, allTasks, "default")
// 	}
// 	return err
// }
//
// func (x *feedParser) addRange(bottom, top int) {
// 	for i := bottom + 1; i < top; i++ {
// 		x.todo = append(x.todo, i)
// 	}
// }
//
// func (x *feedParser) Search(bottom, top int) (err error) {
// 	/*
// 	  def infinite_length(bottom=1, top=-1):
// 	    if bottom == 1 and not item_exists(1): return 0  # Starting edge case
// 	    if bottom == top - 1: return bottom  # Result found! (top doesn’t exist)
// 	    if top < 0:  # Searching forward
// 	      top = bottom << 1  # Base 2 hops
// 	      if item_exists(top):
// 	        top, bottom = -1, top # continue searching forward
// 	    else:  # Binary search between bottom and top
// 	      middle = (bottom + top) // 2
// 	      bottom, top = middle, top if item_exists(middle) else bottom, middle
// 	    return infinite_length(bottom, top)  # Tail recursion!!!
// 	*/
// 	if bottom == top-1 {
// 		x.context.Infof("TOP OF RANGE FOUND! @%d", top)
// 		x.addRange(bottom, top)
// 		return nil
// 	}
// 	var fullStop, isStop bool = false, false
// 	if top < 0 { // Searching forward
// 		top = bottom << 1 // Base 2 hops forward
// 		isStop, fullStop, err = x.isStop(top)
// 		if err != nil {
// 			return err
// 		}
// 		if !isStop {
// 			x.addRange(bottom, top)
// 			top, bottom = -1, top
// 		}
// 	} else { // Binary search between top and bottom
// 		middle := (bottom + top) / 2
// 		isStop, fullStop, err = x.isStop(middle)
// 		if err != nil {
// 			return err
// 		}
// 		if isStop {
// 			top = middle
// 		} else {
// 			x.addRange(bottom, middle)
// 			bottom = middle
// 		}
// 	}
// 	if fullStop {
// 		return nil
// 	}
// 	return x.Search(bottom, top) // TAIL RECURSION!!!
// }
//
// func (x *feedParser) isStop(idx int) (isStop, fullStop bool, err error) {
// 	// Gather posts as necessary
// 	posts, err := x.getAndParseFeed(idx)
// 	if err == ErrFeedParse404 {
// 		x.context.Infof("Reached the end of the feed list (%v)", idx)
// 		return true, false, nil
// 	}
// 	if err != nil {
// 		x.context.Errorf("Error decoding ChiveFeed: %s", err)
// 		return false, false, err
// 	}
//
// 	// Check for Duplicates
// 	count := 0
// 	for _, post := range posts {
// 		id, _, err := guidToInt(post.GUID)
// 		if x.guids[id] || err != nil {
// 			continue
// 		}
// 		count++
// 	}
// 	x.posts = append(x.posts, posts...)
//
// 	// Use store_count info to determine if isStop
// 	isStop = count == 0 || DEBUG
// 	fullStop = len(posts) != count && count > 0
// 	if DEBUG {
// 		isStop = idx > DEPTH
// 		fullStop = idx == DEPTH
// 	}
// 	return
// }
//
// func (x *feedParser) getAndParseFeed(idx int) ([]models.Post, error) {
// 	url := pageURL(idx)
//
// 	// Get Response
// 	x.context.Infof("Parsing index %v (%v)", idx, url)
// 	resp, err := x.client.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != 200 {
// 		if resp.StatusCode == 404 {
// 			return nil, ErrFeedParse404
// 		}
// 		return nil, fmt.Errorf("Feed parcing recieved a %d Status Code", resp.StatusCode)
// 	}
//
// 	// Decode Response
// 	decoder := xml.NewDecoder(resp.Body)
// 	var feed struct {
// 		Items []models.Post `xml:"channel>item"`
// 	}
// 	if decoder.Decode(&feed) != nil {
// 		return nil, err
// 	}
//
// 	// Cleanup Response
// 	for idx := range feed.Items {
// 		post := &feed.Items[idx]
// 		for i, img := range post.Media {
// 			post.Media[i].URL = stripQuery(img.URL)
// 		}
// 		post.MugShot = post.Media[0].URL
// 		post.Media = post.Media[1:]
// 	}
// 	return feed.Items, err
// }
//
// func (x *feedParser) storePosts(dirty []models.Post) (err error) {
// 	var posts []models.Post
// 	var keys []*datastore.Key
// 	for _, post := range dirty {
// 		key, err := x.cleanPost(&post)
// 		if err != nil {
// 			continue
// 		}
// 		posts = append(posts, post)
// 		keys = append(keys, key)
// 	}
// 	if len(keys) > 0 {
// 		complete, err := datastore.PutMulti(x.context, keys, posts)
// 		if err == nil {
// 			err = keycache.AddKeys(x.context, models.POST, complete)
// 		}
// 	}
// 	return err
// }
//
// func (x *feedParser) cleanPost(p *models.Post) (*datastore.Key, error) {
// 	id, link, err := guidToInt(p.GUID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Remove link posts
// 	if link {
// 		x.context.Infof("Ignoring links post %v \"%v\"", p.GUID, p.Title)
// 		return nil, fmt.Errorf("Ignoring links post")
// 	}
//
// 	// Detect video only posts
// 	video := regexp.MustCompile("\\([^&]*Video.*\\)")
// 	if video.MatchString(p.Title) {
// 		x.context.Infof("Ignoring video post %v \"%v\"", p.GUID, p.Title)
// 		return nil, fmt.Errorf("Ignoring video post")
// 	}
// 	x.context.Infof("Storing post %v \"%v\"", p.GUID, p.Title)
//
// 	// Cleanup post titles
// 	clean := regexp.MustCompile("\\W\\(([^\\)]*)\\)$")
// 	p.Title = clean.ReplaceAllLiteralString(p.Title, "")
//
// 	// Post
// 	// temp_key := datastore.NewIncompleteKey(x.context, DB_POST_TABLE, nil)
// 	key := datastore.NewKey(x.context, models.POST, "", id, nil)
// 	return key, nil
// }
//
// func guidToInt(guid string) (int64, bool, error) {
// 	// Remove link posts
// 	url, err := url.Parse(guid)
// 	if err != nil {
// 		return -1, false, err
// 	}
//
// 	// Parsing post id from guid url
// 	id, err := strconv.Atoi(url.Query().Get("p"))
// 	if err != nil {
// 		return -1, false, err
// 	}
// 	return int64(id), url.Query().Get("post_type") == "sdac_links", nil
// }
//
// func stripQuery(dirty string) string {
// 	obj, err := url.Parse(dirty)
// 	if err != nil {
// 		return dirty
// 	}
// 	obj.RawQuery = ""
// 	return obj.String()
// }
