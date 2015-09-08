package crawler

import (
  // "app/models"
  // "app/helpers/keycache"
  "appengine"
  // "appengine/datastore"
  // "appengine/delay"
  // "appengine/taskqueue"
  "appengine/urlfetch"
  // "encoding/xml"
  // "fmt"
  "net/http"
  "strconv"
)

var (
  DEBUG = true
  DEBUG_DEPTH = 1
)

func NewFeedCrawler(c appengine.Context) *FeedCrawler {
  return &FeedCrawler{
    context: c,
    client:  urlfetch.Client(c),
    results: make(chan ChivePost),
  }
}

type FeedCrawler struct {
  context appengine.Context
  client  *http.Client

  todo    []int
  guids   map[string]bool // this could be extremely large
  results chan ChivePost
}

func (fc *FeedCrawler) StartSearch() <-chan ChivePost {
  go func() {
    defer close(fc.results)
    for i := 0; i < 99; i++ {
      fc.results <- ChivePost{KEY:"asdf", XML:strconv.Itoa(i)}
    }
    // fc.search(1, -1)
  }()
  return fc.results
}

func (fc *FeedCrawler) addRange(bot, top int) {
  // TODO: isn't there a better way to perform this operation!?
  for i := bot + 1; i < top; i++ {
    fc.todo = append(fc.todo, i)
  }
}

// func (fc *FeedCrawler) search(bot, top int) (err error) {
//   /*
//   def infinite_length(bottom=1, top=-1):
//     if bottom == 1 and not item_exists(1): return 0  # Starting edge case
//     if bottom == top - 1: return bottom  # Result found! (top doesn’t exist)
//     if top < 0:  # Searching forward
//       top = bottom << 1  # Base 2 hops
//       if item_exists(top):
//         top, bottom = -1, top # continue searching forward
//     else:  # Binary search between bottom and top
//       middle = (bottom + top) // 2
//       bottom, top = middle, top if item_exists(middle) else bottom, middle
//     return infinite_length(bottom, top)  # Tail recursion!!!
//   */
//   if bot == top - 1 {
//     fc.context.Infof("TOP OF RANGE FOUND! @%d", top)
//     fc.addRange(bot, top)
//     return nil
//   }
//   var full_stop, is_stop bool = false, false
//   if top < 0 { // Searching forward
//     top = bot << 1  // Base 2 hops forward
//     is_stop, full_stop, err = fc.isStop(top)
//     if err != nil {
//       return err
//     }
//     if !is_stop {
//       fc.addRange(bot, top)
//       top, bot = -1, top
//     }
//   } else { // Binary search between top and bottom
//     mid := (bot + top) / 2
//     is_stop, full_stop, err = fc.isStop(mid)
//     if err != nil {
//       return err
//     }
//     if is_stop {
//       top = mid
//     } else {
//       fc.addRange(bot, mid)
//       bot = mid
//     }
//   }
//   if full_stop {
//     return nil
//   }
//   return fc.search(bot, top)  // TAIL RECURSION!!!
// }
//
// func (fc *FeedCrawler) isStop(idx int) (is_stop, full_stop bool, err error) {
//   // Gather posts as necessary
//   posts, err := fc.getAndParseFeed(idx)
//   if err == FeedParse404Error {
//     fc.context.Infof("Reached the end of the feed list (%v)", idx)
//     return true, false, nil
//   }
//   if err != nil {
//     fc.context.Errorf("Error decoding ChiveFeed: %s", err)
//     return false, false, err
//   }
//
//   // Check for Duplicates
//   store_count := 0
//   for _, post := range posts {
//     id, _, err := guidToInt(post.Guid)
//     if x.guids[id] || err != nil {
//       continue
//     }
//     store_count += 1
//   }
//   fc.posts = append(fc.posts, posts...)
//
//   // Use store_count info to determine if isStop
//   is_stop = store_count == 0 || DEBUG
//   full_stop = len(posts) != store_count && store_count > 0
//   if DEBUG {
//     is_stop = idx > DEBUG_DEPTH
//     full_stop = idx == DEBUG_DEPTH
//   }
//   return
// }
