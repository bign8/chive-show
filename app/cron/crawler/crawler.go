package crawler

import (
	// "app/models"
	// "app/helpers/keycache"
	"appengine"
	// "appengine/datastore"
	// "appengine/delay"
	// "appengine/taskqueue"

	"fmt"
	"net/http"
)

const (
	// DEBUG enable if troubleshooting algorithm
	DEBUG = false

	// DEPTH depth of feed mining
	DEPTH = 1

	// XML name of where xml posts are stored
	XML = "xml"
)

type Data struct {
	KEY string
	XML string
}

func Crawl(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	// fetcher, dePager, parser, batcher, saver
	pages := Fetcher(c)
	posts := UnPager(c, pages)
	batch := Batcher(posts, 50)
	Storage(c, batch)
	fmt.Fprint(w, "Crawl Complete!")
}
