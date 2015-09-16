package cron

import (
  "github.com/bign8/chive-show/app/models"
  "github.com/bign8/chive-show/app/helpers/keycache"
  "appengine"
  "appengine/datastore"
  "appengine/delay"
  "appengine/taskqueue"
  "appengine/urlfetch"
  "encoding/xml"
  "fmt"
  "github.com/mjibson/appstats"
  "net/http"
  "net/url"
  "regexp"
  "strconv"
)

const (
  TODO_BATCH_SIZE = 10
  DEBUG = true
  DEBUG_DEPTH = 1
  PROCESS_TODO_DEFERRED = true
)

func Init() {
  http.Handle("/cron/parse", appstats.NewHandler(parseFeeds))
  http.HandleFunc("/cron/delete", delete)
}

var (
  FeedParse404Error error = fmt.Errorf("Feed parcing recieved a %d Status Code", 404)
)

func page_url(idx int) string {
  return fmt.Sprintf("http://thechive.com/feed/?paged=%d", idx)
}

func parseFeeds(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
  guids    map[int64]bool  // this could be extremely large
  posts    []models.Post
}

func (x *FeedParser) Main(c appengine.Context, w http.ResponseWriter) error {
  x.context = c
  x.client = urlfetch.Client(c)

  // Load guids from DB
  // TODO: do this with sharded keys
  keys, err := datastore.NewQuery(models.DB_POST_TABLE).KeysOnly().GetAll(c, nil)
  if err != nil {
    c.Errorf("Error finding keys %v %v", err, appengine.IsOverQuota(err))
    return err
  }
  x.guids = map[int64]bool{}
  for _, key := range keys {
    x.guids[key.IntID()] = true
  }
  keys = nil

  // // DEBUG ONLY
  // data, err := json.MarshalIndent(x.guids, "", "  ")
  // fmt.Fprint(w, string(data))
  // return err
  x.posts = make([]models.Post, 0)

  // Initial recursive edge case
  is_stop, full_stop, err := x.isStop(1)
  if is_stop || full_stop || err != nil {
    c.Infof("Finished without recursive searching %v", err)
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
        err = x.storePosts(posts)
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
  if PROCESS_TODO_DEFERRED && len(all_tasks) > 0 {
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

  // Check for Duplicates
  store_count := 0
  for _, post := range posts {
    id, _, err := guidToInt(post.Guid)
    if x.guids[id] || err != nil {
      continue
    }
    store_count += 1
  }
  x.posts = append(x.posts, posts...)

  // Use store_count info to determine if isStop
  is_stop = store_count == 0 || DEBUG
  full_stop = len(posts) != store_count && store_count > 0
  if DEBUG {
    is_stop = idx > DEBUG_DEPTH
    full_stop = idx == DEBUG_DEPTH
  }
  return
}

func (x *FeedParser) getAndParseFeed(idx int) ([]models.Post, error) {
  url := page_url(idx)

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
    Items []models.Post `xml:"channel>item"`
  }
  if decoder.Decode(&feed) != nil {
    return nil, err
  }

  // Cleanup Response
  for idx := range feed.Items {
    post := &feed.Items[idx]
    for i, img := range post.Media {
      post.Media[i].Url = stripQuery(img.Url)
    }
    post.MugShot = post.Media[0].Url
    post.Media = post.Media[1:]
  }
  return feed.Items, err
}

func (x *FeedParser) storePosts(dirty_posts []models.Post) (err error) {
  posts := make([]models.Post, 0)
  keys := make([]*datastore.Key, 0)
  for _, post := range dirty_posts {
    key, err := x.cleanPost(&post)
    if err != nil {
      continue
    }
    posts = append(posts, post)
    keys = append(keys, key)
  }
  if len(keys) > 0 {
    complete_keys, err := datastore.PutMulti(x.context, keys, posts)
    if err == nil {
      err = keycache.AddKeys(x.context, models.DB_POST_TABLE, complete_keys)
    }
  }
  return err
}

func (x *FeedParser) cleanPost(p *models.Post) (*datastore.Key, error) {
  id, is_link_post, err := guidToInt(p.Guid)
  if err != nil {
    return nil, err
  }
  // Remove link posts
  if is_link_post {
    x.context.Infof("Ignoring links post %v \"%v\"", p.Guid, p.Title)
    return nil, fmt.Errorf("Ignoring links post")
  }

  // Detect video only posts
  video_re := regexp.MustCompile("\\([^&]*Video.*\\)")
  if video_re.MatchString(p.Title) {
    x.context.Infof("Ignoring video post %v \"%v\"", p.Guid, p.Title)
    return nil, fmt.Errorf("Ignoring video post")
  }
  x.context.Infof("Storing post %v \"%v\"", p.Guid, p.Title)

  // Cleanup post titles
  clean_re := regexp.MustCompile("\\W\\(([^\\)]*)\\)$")
  p.Title = clean_re.ReplaceAllLiteralString(p.Title, "")

  // Post
  // temp_key := datastore.NewIncompleteKey(x.context, DB_POST_TABLE, nil)
  temp_key := datastore.NewKey(x.context, models.DB_POST_TABLE, "", id, nil)
  return temp_key, nil
}

func guidToInt(guid string) (int64, bool, error) {
  // Remove link posts
  url, err := url.Parse(guid)
  if err != nil {
    return -1, false, err
  }

  // Parsing post id from guid url
  temp_id, err := strconv.Atoi(url.Query().Get("p"))
  if err != nil {
    return -1, false, err
  }
  return int64(temp_id), url.Query().Get("post_type") == "sdac_links", nil
}

func stripQuery(dirty_url string) string {
  obj, err := url.Parse(dirty_url)
  if err != nil {
    return dirty_url
  }
  obj.RawQuery = ""
  return obj.String()
}
