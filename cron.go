package main

import (
  "appengine"
  "appengine/datastore"
  "appengine/delay"
  "appengine/taskqueue"
  "appengine/urlfetch"
  "encoding/xml"
  "encoding/json"
  "fmt"
  "github.com/mjibson/appstats"
  "net/http"
)

const (
  TODO_BATCH_SIZE = 10
  DEBUG = false
  DEBUG_DEPTH = 16
  PROCESS_TODO_DEFERRED = true
)

func cron() {
  // http.HandleFunc("/cron/parse_feeds", parseFeeds)
  http.Handle("/cron/parse", appstats.NewHandler(parseFeeds))
  http.HandleFunc("/cron/delete", func(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    q := datastore.NewQuery("Post").KeysOnly()
    keys, err := q.GetAll(c, nil)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    // Batch Delete
    var del_keys []*datastore.Key
    for _, key := range keys {
      if del_keys == nil {
        del_keys = make([]*datastore.Key, 0)
      }
      del_keys = append(del_keys, key)
      if len(del_keys) > 50 {
        err = datastore.DeleteMulti(c, del_keys)
        c.Infof("Deleting Keys %v", del_keys)
        del_keys = nil
      }
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    }
    if len(del_keys) > 0 {
      datastore.DeleteMulti(c, del_keys)
    }
    fmt.Fprint(w, "DELETED")
  })
}

var FeedParse404Error error = fmt.Errorf("Feed parcing recieved a %d Status Code", 404)

func page_url(idx int) string {
  if idx > 0 {
    return fmt.Sprintf("http://thechive.com/feed/?paged=%d", idx - 1)
  } else {
    return fmt.Sprintf("http://thechive.com/feed/?page")
  }
}

func parseFeeds(c appengine.Context, w http.ResponseWriter, r *http.Request) {
  // c := appengine.NewContext(r)
  fp := new(FeedParser)
  err := fp.Main(c, w)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    fmt.Fprint(w, "Parsed")
  }
}

type FeedParser struct {
  context  appengine.Context
  client   *http.Client

  todo     []int
  guids    map[string]bool  // this could be extremely large
}

func (x *FeedParser) Main(c appengine.Context, w http.ResponseWriter) error {
  x.context = c
  x.client = urlfetch.Client(c)

  // Load guids from DB
  var posts []Post
  q := datastore.NewQuery("Post") //.Project("guid", "link") // TODO: Figure out projection queries
  if _, err := q.GetAll(c, &posts); err != nil {
    c.Errorf("Error projecting %v %v", err)
    return err
  }
  x.guids = map[string]bool{}
  for _, post := range posts {
    x.guids[post.Guid] = true
  }
  posts = nil

  // // DEBUG ONLY
  // data, err := json.MarshalIndent(x.guids, "", "  ")
  // fmt.Fprint(w, string(data))
  // return err

  // Initial recursive edge case
  is_stop, full_stop, err := x.isStop(1)
  if is_stop || full_stop || err != nil {
    c.Infof("Finished without recursive searching %v", err)
    return err
  }

  // Recursive search strategy
  err = x.Search(1, -1)
  if err == nil {
    err = x.processTodo()
  }
  if err != nil {
    c.Errorf("Error in Main %v", err)
  }
  return err
}

var processBatchDeferred = delay.Func("process-todo-batch", func(c appengine.Context, ids []int) {
  parser := FeedParser{
    context: c,
    client: urlfetch.Client(c),
  }
  parser.processBatch(ids)
})

func (x *FeedParser) processBatch(ids []int) error {
  done := make(chan error)
  for _, idx := range ids {
    go func (idx int) {
      posts, err := x.getAndParseFeed(idx)
      if err == nil {
        for _, post := range posts {
          err = x.storePost(post)
          if err != nil {
            break
          }
        }
      }
      done <- err
    }(idx)
  }
  for i := 0; i < len(ids); i++ {
    err := <-done
    if err != nil {
      x.context.Errorf("error storing feed (at index %d): %v", i, err)
      return err
    }
  }
  return nil
}

func (x *FeedParser) processTodo() error {
  x.context.Infof("Processing TODO: %v", x.todo)

  var batch []int
  var task *taskqueue.Task
  all_tasks := make([]*taskqueue.Task, 0)
  var err error
  for _, idx := range x.todo {
    if batch == nil {
      batch = make([]int, 0)
    }
    batch = append(batch, idx)
    if len(batch) >= TODO_BATCH_SIZE {
      if PROCESS_TODO_DEFERRED {
        task, err = processBatchDeferred.Task(batch)
        if err == nil {
          all_tasks = append(all_tasks, task)
        }
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
    if PROCESS_TODO_DEFERRED {
      task, err = processBatchDeferred.Task(batch)
      if err == nil {
        all_tasks = append(all_tasks, task)
      }
    } else {
      err = x.processBatch(batch)
    }
  }
  if PROCESS_TODO_DEFERRED {
    x.context.Infof("Adding %d task(s) to the default queue", len(all_tasks))
    taskqueue.AddMulti(x.context, all_tasks, "default")
  }
  return err
}

func (x *FeedParser) addRange(bottom, top int) {
  for i := bottom + 1; i < top; i++ {
    x.todo = append(x.todo, i)
  }
}

func (x *FeedParser) Search(bottom, top int) (err error) {
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
  if bottom == top - 1 {
    x.context.Infof("TOP OF RANGE FOUND! @%d", top)
    x.addRange(bottom, top)
    return nil
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
    middle := (bottom + top) / 2
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
  return x.Search(bottom, top)  // TAIL RECURSION!!!
}

func (x *FeedParser) isStop(idx int) (is_stop, full_stop bool, err error) {
  // Gather posts as necessary
  posts, err := x.getAndParseFeed(idx)
  if err == FeedParse404Error {
    x.context.Infof("Reached the end of the feed list (%v)", idx)
    return true, false, nil
  }
  if err != nil {
    x.context.Errorf("Error decoding ChiveFeed: %s", err)
    return false, false, err
  }

  // Iterate posts, store store as necessary
  // TODO: make storePost into a go routine, use error channel for callbacks
  store_count := 0
  for _, post := range posts {
    if x.guids[post.Guid] {
      continue
    }
    if err = x.storePost(post); err != nil {
      x.context.Errorf("Error in storePost %v", err)
      return false, false, err
    }
    store_count += 1
  }
  x.context.Infof("Stored %d posts for feed %d", store_count, idx)

  // Use storePost info to determine if isStop
  is_stop = store_count == 0 || DEBUG
  full_stop = len(posts) != store_count && store_count > 0
  if DEBUG {
    is_stop = idx > DEBUG_DEPTH
    full_stop = idx == DEBUG_DEPTH
  }
  return
}

func (x *FeedParser) getAndParseFeed(idx int) ([]Post, error) {
  url := page_url(idx - 1)

  // Get Response
  x.context.Infof("Parsing index %v (%v)", idx, url)
  resp, err := x.client.Get(url)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    if resp.StatusCode == 404 {
      return nil, FeedParse404Error
    }
    return nil, fmt.Errorf("Feed parcing recieved a %d Status Code", resp.StatusCode)
  }

  // Decode Response
  decoder := xml.NewDecoder(resp.Body)
  var feed struct {
    Items []Post `xml:"channel>item"`
  }
  if decoder.Decode(&feed) != nil {
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

func (x *FeedParser) storePost(p Post) (err error) {
  x.context.Infof("Storing post %v \"%v\"", p.Guid, p.Title)

  // Creator
  // temp_key := datastore.NewIncompleteKey(x.context, "Author", nil)
  // creator_key, err := datastore.Put(x.context, temp_key, &p.JsCreator)
  p.Creator, err = json.Marshal(&p.JsCreator)
  if err != nil {
    x.context.Errorf("Error storePost Marshal 1 %v", err)
    return err
  }

  // Media
  // TODO: store imgs
  p.Media, err = json.Marshal(&p.JsImgs)
  if err != nil {
    x.context.Errorf("Error storePost Marshal 2 %v", err)
    return err
  }

  // Post
  temp_key := datastore.NewIncompleteKey(x.context, "Post", nil)
  _, err = datastore.Put(x.context, temp_key, &p)
  return err
}
