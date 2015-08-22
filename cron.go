package main

import (
  // "encoding/json"
  "fmt"
  // "math/rand"
  "net/http"
  // "net/url"
  // "strconv"
  //
  // "helpers"
  //
  "appengine"
  "appengine/urlfetch"
  "encoding/xml"
  "encoding/json"
  "errors"
  // "appengine/datastore"

)

func cron() {
  http.HandleFunc("/cron/parse_feeds", parseFeeds)
}

func page_url(idx int) string {
  if idx > 0 {
    return fmt.Sprintf("http://thechive.com/feed/?paged=%d", idx - 1)
  } else {
    return fmt.Sprintf("http://thechive.com/feed/?page")
  }
}

func parseFeeds(w http.ResponseWriter, r *http.Request) {
  // fmt.Fprint(w, "Here is your cron\n")
  // fmt.Fprint(w, page_url(0))
  // fmt.Fprint(w, page_url(3))

  c := appengine.NewContext(r)
  fp := new(FeedParser)
  fp.Main(c, w)
  c.Infof("TODO: %v", fp.todo)
}

type FeedParser struct {
  context  appengine.Context
  client   *http.Client
  writer   http.ResponseWriter

  todo     []int
}

func (x *FeedParser) Main(c appengine.Context, w http.ResponseWriter) error {
  x.context = c
  x.writer = w
  x.client = urlfetch.Client(c)
  is_stop, full_stop, err := x.isStop(1)
  if is_stop || full_stop || err != nil {
    return err
  }
  err = x.Search(1, -1)
  if err != nil {
    for _, idx := range x.todo {
      go x.pullNstore(idx)
    }
    // TODO: process x.todo
  }
  return err
}

func (x *FeedParser) pullNstore(idx int) (err error) {
  url := page_url(idx - 1)
  resp, err := x.client.Get(url)
  if err != nil {
    return err
  }
  // TODO: process and store feed here
  x.context.Infof("Content of %v: %v", url, resp)
  return nil
}

func (x *FeedParser) addRange(bottom, top int) {
  for i := bottom + 1; i < top; i++ {
    x.todo = append(x.todo, i)
  }
}

func (x *FeedParser) Search(bottom, top int) (err error) {
  if bottom == top - 1 {
    x.addRange(bottom, top)
    return nil // Todo, store this
  }
  var full_stop, is_stop bool = false, false
  if top < 0 { // Searching forward
    top = bottom << 1  // Base 2 hops forward
    is_stop, full_stop, err = x.isStop(top)
    if err != nil {
      return err
    }
    if !is_stop {
      x.addRange(bottom, top)
      top, bottom = -1, top
    }
  } else { // Binary search between top and bottom
    middle := (bottom + top) / 2  // make sure int
    is_stop, full_stop, err = x.isStop(middle)
    if err != nil {
      return err
    }
    if is_stop {
      top = middle
    } else {
      x.addRange(bottom, middle)
      bottom = middle
    }
  }
  if full_stop {
    return nil
  }
  return x.Search(bottom, top)
}

func (x *FeedParser) isStop(idx int) (is_stop, full_stop bool, err error) {
  url := page_url(idx - 1)
  resp, err := x.client.Get(url)
  if err != nil {
    x.context.Errorf("Error retrieving \"%v\": %v", url, err)
    return false, false, err
  }
  defer resp.Body.Close()
  if resp.StatusCode == 404 {
    x.context.Infof(" the end of the feed list %v (%v)", )
    return true, false, nil
  }
  if resp.StatusCode != 200 {
    return false, false, errors.New("Recieved non-200/404 error code")
  }

  // Convert the xml to struct
  posts, err := parseFeedResponse(resp)
  if err != nil {
    x.context.Errorf("Error decoding ChiveFeed: %s", err)
    return false, false, err
  }

  data, err := json.MarshalIndent(posts, "", " ")
  fmt.Fprint(x.writer, string(data))

  // Actually loading up
  x.context.Infof("Loading index %v %v", idx, url)
  top := 1
  return idx > top, idx == top, nil

  // // TODO: process feed here, store if necessary
  // // full_stop is if this is the final collision (ie, some new posts, but not all)
  // x.context.Infof("Content of %v: %v", url, resp)
  // return false, false
}

func parseFeedResponse(resp *http.Response) ([]Post, error) {
  // Decode Response
  decoder := xml.NewDecoder(resp.Body)
  var feed ChiveFeed
  err := decoder.Decode(&feed)
  if err != nil {
    return nil, err
  }

  // Cleanup Response
  for idx := range feed.Items {
    post := &feed.Items[idx]
    post.JsCreator = Author{
      Name: post.StrAuthor,
      Img: post.JsImgs[0].Url,
    }
    post.JsImgs = post.JsImgs[1:]
  }
  return feed.Items, err
}

type ChiveFeed struct {
  Items []Post `xml:"channel>item"`
}

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
