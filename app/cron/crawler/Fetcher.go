package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"appengine"
	"appengine/urlfetch"
)

const (
	// DEBUG enable if troubleshooting algorithm
	DEBUG = true

	// DEPTH depth of feed mining
	DEPTH = 1
)

func pageURL(idx int) string {
	return fmt.Sprintf("http://thechive.com/feed/?paged=%d", idx)
}

// Fetcher returns stream of un-processed xml posts
func Fetcher(c appengine.Context) <-chan string {
	res := make(chan string)
	worker := &fetcher{
		res:     res,
		context: c,
		client:  urlfetch.Client(c),
	}
	go worker.Main()
	return res
}

type fetcher struct {
	res     chan<- string
	context appengine.Context
	client  *http.Client
	todo    chan int
}

func (x *fetcher) Main() error {
	defer close(x.res)

	// Check first item edge case
	if isStop, err := x.isStop(1); isStop || err != nil {
		x.context.Infof("Fetcher: Finished without recursive searching %v", err)
		return err
	}

	// Defer as many todo workers as necessary
	x.todo = make(chan int)
	defer close(x.todo)
	go x.processTODO()
	return x.Search(1, -1)
}

func (x *fetcher) Search(bottom, top int) (err error) {
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
		x.context.Infof("Fetcher: TOP OF RANGE FOUND! @%d", top)
		x.addRange(bottom, top)
		return nil
	}
	x.context.Infof("Fetcher: Search(%d, %d)", bottom, top)
	var isStop = false

	// Searching forward
	if top < 0 {
		top = bottom << 1 // Base 2 hops forward
		isStop, err = x.isStop(top)
		if err != nil {
			return err
		}
		if !isStop {
			x.addRange(bottom, top)
			top, bottom = -1, top
		}

		// Binary search between top and bottom
	} else {
		middle := (bottom + top) / 2
		isStop, err = x.isStop(middle)
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
	return x.Search(bottom, top) // TAIL RECURSION!!!
}

func (x *fetcher) isStop(idx int) (isStop bool, err error) {

	// Gather posts as necessary
	url := pageURL(idx)
	x.context.Infof("Fetcher: Fetching %s", url)
	resp, err := x.client.Get(url)
	if err != nil {
		x.context.Errorf("Fetcher: Error decoding ChiveFeed: %s", err)
		return true, err
	}
	defer resp.Body.Close()

	// Check Response Codes for non-200 responses
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			x.context.Infof("Fetcher: Reached the end of the feed list (%v)", idx)
			return true, nil
		}
		return true, fmt.Errorf("Fetcher: Feed parcing received a %d Status Code on (%s)", resp.StatusCode, url)
	}

	// Pull response content into String
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, err
	}
	x.res <- string(contents)

	// Use store_count info to determine if isStop
	if DEBUG {
		isStop = idx >= DEPTH
	}
	return isStop, nil
}

func (x *fetcher) addRange(bottom, top int) {
	for i := bottom + 1; i < top; i++ {
		x.todo <- i
	}
}

func (x *fetcher) processTODO() {
	for idx := range x.todo {
		x.context.Infof("Fetcher: NOT processing TODO %d", idx)
		//x.isStop(idx)
	}
}
