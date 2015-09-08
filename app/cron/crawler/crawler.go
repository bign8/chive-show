package crawler

import (
  // "app/models"
  // "app/helpers/keycache"
  "appengine"
  // "appengine/datastore"
  // "appengine/delay"
  // "appengine/taskqueue"
  "appengine/urlfetch"
  "encoding/xml"
  "fmt"
  "net/http"
)

// Sourcer: this is a source for defered work chains

type ChivePost struct {
  KEY string `xml:"guid"`
  XML string `xml:",innerxml"`
}

type ChivePostMiner struct {
  Item ChivePost `xml:"channel>item"`
}

func page_url(idx int) string {
  return fmt.Sprintf("http://thechive.com/feed/?paged=%d", idx)
}


func Crawl(c appengine.Context, w http.ResponseWriter, r *http.Request) {
  url := page_url(0)

  // Get Response
  c.Infof("Parsing index 0 (%v)", url)
  resp, err := urlfetch.Client(c).Get(url)
  if err != nil {
    fmt.Fprint(w, "client error")
    return
  }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    fmt.Fprint(w, "unexpected error code")
  }

  // Decode Response
  var feed []ChivePostMiner
  decoder := xml.NewDecoder(resp.Body)
  if err := decoder.Decode(&feed); err != nil {
    c.Errorf("decode error %v", err)
    fmt.Fprint(w, "decode error")
    return
  }

  feed[0].Item.XML = "<item>" + feed[0].Item.XML + "</item>"

  c.Infof("Something %v", feed)

  // TODO: store all items to datastore


  // DEBUGGING ONLY.... HERE DOWN

  // post, err := parseData(feed[0].Item.XML)
  // if err != nil {
  //   c.Errorf("error parsing %v", err)
  //   return
  // }
  //
  // // JSONIFY Response
  // str_items, err := json.MarshalIndent(&post, "", "  ")
  // var out string
  // if err != nil {
  //   out = "{\"status\":\"error\",\"code\":500,\"data\":null,\"msg\":\"Error marshaling data\"}"
  // } else {
  //   out = string(str_items)
  // }
  // w.Header().Set("Content-Type", "application/json; charset=utf-8")
  // fmt.Fprint(w, out)
}

func Crawl2(c appengine.Context, w http.ResponseWriter, r *http.Request) {
  crawler := NewFeedCrawler(c)
  found_posts := crawler.StartSearch()
  batch_posts := Batcher(found_posts, 20)
  Storage(batch_posts, c)
}

func Storage(in <-chan []ChivePost, c appengine.Context) {
  go func() {
    for batch := range in {
      fmt.Println(batch)
      c.Infof("Storing %v", batch)
    }
  }()
}
