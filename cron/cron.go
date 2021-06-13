package cron

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/anaskhan96/soup"
	"go.opencensus.io/plugin/ochttp"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/bign8/chive-show/appengine"
	"github.com/bign8/chive-show/models"
	"github.com/bign8/chive-show/models/datastore"
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

// TODO: use a client that supports gzip request/responses
var client = &http.Client{
	Transport: &ochttp.Transport{
		// Use Google Cloud propagation format.
		Propagation: &propagation.HTTPFormat{},
	},
}

// Init initializes cron handlers
func Init(store *datastore.Store) {
	tasker, err := cloudtasks.NewClient(context.Background())
	if err != nil {
		panic(err)
	}
	http.Handle("/cron/parse", parse(store, tasker))
	http.Handle("/cron/batch", batch(store))
	http.HandleFunc("/cron/debug", debug)
	http.HandleFunc("/cron/test", test)

	http.Handle("/cron/rebuild", RebuildHandler(store)) // Step 0: rebuild from nothing (on project init)
	http.Handle("/cron/crawl", CrawlHandler(store))     // Step 1: search for posts (hourly)
	http.Handle("/cron/mine", MineHandler(store))       // Step 2: load post metadata
}

func debug(w http.ResponseWriter, r *http.Request) {
	bits, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	w.Write(bits)
}

var (
	// ErrFeedParse404 if feed page is not found
	ErrFeedParse404 = fmt.Errorf("feed parcing recieved a %d Status Code", 404)
)

func pageURL(idx int) string {
	if idx == 1 {
		return "https://thechive.com/feed/"
	}
	return fmt.Sprintf("https://thechive.com/feed/?paged=%d", idx)
}

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
			_, posts, err := x.getAndParseFeed(idx)
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
	found, posts, err := x.getAndParseFeed(idx)
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

func test(w http.ResponseWriter, r *http.Request) {
	x := &feedParser{
		context: r.Context(),
	}
	// _, posts, err := x.getAndParseFeed(1)
	post := models.Post{
		Link: "https://thechive.com/2021/06/06/all-they-had-to-do-was-change-a-websites-phone-number-to-avoid-this-revenge/",
	}
	err := x.mine(&post)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent(``, ` `)
	enc.SetEscapeHTML(false)
	enc.Encode(post)
}

func (x *feedParser) getAndParseFeed(idx int) (found int, posts []models.Post, err error) {
	url := pageURL(idx)

	// Get Response
	log.Printf("INFO: Parsing index %v (%v)", idx, url)
	body, err := x.fetch(url)
	if err != nil {
		return 0, nil, err
	}
	defer body.Close()

	// Decode Response
	decoder := xml.NewDecoder(body)
	var feed struct {
		Items []models.Post `xml:"channel>item"`
	}
	if err := decoder.Decode(&feed); err != nil {
		return 0, nil, err
	}

	// Remove undesired posts to reduce the amount of mined data and trash in the database
	//  - dopamine dump posts (they link out to i.thechive.com and host external user content)
	//  - duplicates!
	remove := func(i int) {
		feed.Items = append(feed.Items[:i], feed.Items[i+1:]...)
	}
	found = len(feed.Items)
	for i := len(feed.Items) - 1; i >= 0; i-- {
		post := feed.Items[i]
		if post.Link == "http://i.thechive.com/dopamine-dump" {
			// https://i.thechive.com/rest/uploads?queryType=dopamine-dump&offset=0
			log.Printf("INFO: Ignoring dopamine dump (TODO: separate miner for i.thechive.com/rest/uploads): %s", post.Link)
			remove(i)
		} else if has, err := x.store.Has(x.context, post); has || err != nil {
			log.Printf("INFO: Found duplicate: %s %v", post.Link, err)
			remove(i)
		}
	}

	// Cleanup Response
	// TODO: mine pages in parallel (worker pool?)
	for idx := range feed.Items {
		post := &feed.Items[idx]
		if err = x.mine(post); err != nil {
			log.Printf("Unable to mine page for details: %s", post.Link)
			return found, nil, err
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
	return found, feed.Items, nil
}

// NOTE: there is an alternate link in the header of posts that points to a JSON representation of the post
// ex: https://thechive.com/?p=3683573 => https://thechive.com/wp-json/wp/v2/posts/3683573
func (x *feedParser) mine(post *models.Post) error {
	body, err := x.fetch(post.Link)
	if err != nil {
		log.Printf("mine(%s): unable to fetch: %s", post.Link, err)
		return nil
	}
	dom, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err = body.Close(); err != nil {
		return err
	}
	doc := soup.HTMLParse(string(dom))

	// Pages embed single banner image
	figure := doc.Find("figure")
	if figure.Error != nil {
		log.Printf("WARNING: unable to load figure %s: %v", post.Link, figure.Error)
		return nil
	}
	obj := figure.Find("img")
	if obj.Error != nil {
		obj = figure.Find("source")
	}
	if obj.Error != nil {
		log.Printf("WARNING: uanble to load banner content %s: %v", post.Link, obj.Error)
	} else {
		media := models.Media{URL: obj.Attrs()["src"]}

		// Attempt to scrape captions as well
		caption := figure.Find("figcaption", "class", "gallery-caption")
		if caption.Error == nil {
			for _, ele := range caption.Children() {
				media.Caption += ele.HTML()
			}
			media.Caption = strings.TrimSpace(media.Caption)
		}

		post.Media = append(post.Media, media)
	}

	// parse CHIVE_GALLERY_ITEMS from script id='chive-theme-js-js-extra' into JSON
	// TODO: use match the image prefix? "https:\/\/thechive.com\/wp-content\/uploads\/" in the HTML and parse to closing "
	js := doc.Find("script", "id", "chive-theme-js-js-extra")
	if js.Error != nil {
		log.Printf("WARNING: Unable to find script logic in %q %v", post.Link, js.Error)
		return nil
	}
	src := js.FullText()
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
			HTML string `json:"html"`
			Type string `json:"type"`
		} `json:"items"`
	}
	err = json.Unmarshal([]byte(src), &what)
	if err != nil {
		panic(err)
	}

	// Parse HTML attributes of JSON to get images
	for i, obj := range what.Items {
		if obj.HTML == "" {
			// first entry allways appears to be empty as the first post is embedded in page content
			// that said, this warning is here in case something changes in the future
			if i != 0 {
				log.Printf("WARNING: got nil HTML in item %d on %s", i, post.Link)
			}
			continue
		}
		ele := soup.HTMLParse(obj.HTML)
		if ele.Error != nil {
			log.Printf("WARNING: unable to parse HTML of %d on %s", i, post.Link)
			continue
		}
		var imgs []soup.Root
		switch obj.Type {
		case "gif":
			imgs = ele.FindAll("source")
		default:
			imgs = ele.FindAll("img")
		}
		if len(imgs) == 0 {
			log.Printf("WARNING: No media found in item %d on %s", i, post.Link)
			continue
		}
		for _, img := range imgs {
			media := models.Media{URL: img.Attrs()["src"]}

			// Attempt to scrape captions as well
			caption := ele.Find("figcaption", "class", "gallery-caption")
			if caption.Error == nil {
				for _, ele := range caption.Children() {
					media.Caption += ele.HTML()
				}
				media.Caption = strings.TrimSpace(media.Caption)
			}

			post.Media = append(post.Media, media)
		}
	}
	return nil
}

func (x *feedParser) fetch(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req.WithContext(x.context))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, ErrFeedParse404
		}
		return nil, fmt.Errorf("feed parcing recieved a %d Status Code", resp.StatusCode)
	}
	return resp.Body, nil
}

func stripQuery(dirty string) string {
	obj, err := url.Parse(dirty)
	if err != nil {
		return dirty
	}
	obj.RawQuery = ""
	return obj.String()
}
