package crawler

import (
	// "app/models"
	// "app/helpers/keycache"
	"appengine"
	// "appengine/datastore"
	// "appengine/delay"
	// "appengine/taskqueue"
	"encoding/xml"
	"fmt"
	"net/http"

	"appengine/urlfetch"
)

// Sourcer: this is a source for defered work chains

type chivePost struct {
	KEY string `xml:"guid"`
	XML string `xml:",innerxml"`
}

type chivePostMiner struct {
	Item []chivePost `xml:"channel>item"`
}

func Crawl(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	url := pageURL(1)

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
	var feed chivePostMiner
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&feed); err != nil {
		c.Errorf("decode error %v", err)
		fmt.Fprint(w, "decode error")
		return
	}

	// Wrap posts in xml
	for idx, post := range feed.Item {
		feed.Item[idx].XML = "<item>" + post.XML + "</item>"
	}

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
	// fetcher, dePager, parser, batcher, saver
	pages := Fetcher(c)
	posts := UnPager(c, pages)
	for post := range posts {
		c.Infof("Post: %v", post)
	}
	// batch := Batcher(posts, 20)
	// Storage(c, batch)
}

func Storage(c appengine.Context, in <-chan []string) {
	go func() {
		for batch := range in {
			fmt.Println(batch)
			c.Infof("Storing %v", batch)
		}
	}()
}
