package crawler

import (
	// "app/models"
	// "app/helpers/keycache"

	"appengine"
	"appengine/datastore"
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

func Crawl(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	fetchers, storers := 50, 20

	// fetcher, dePager, parser, batcher, saver
	pages := Fetcher(c, fetchers)
	// posts := UnPager(c, pages, pagers)
	batch := Batcher(c, pages, 10)
	Storage(c, batch, storers)

	fmt.Fprint(w, "Crawl Complete!")
}

func Stats(c appengine.Context, w http.ResponseWriter, r *http.Request) {

	q := datastore.NewQuery("xml")

	var data []Store
	keys, err := q.GetAll(c, &data)
	if err != nil {
		fmt.Fprintf(w, "Error %s", err)
		return
	}

	for idx, key := range keys {
		fmt.Fprintf(w, "Data %s: len %d\n", key, len(data[idx].XML))
	}

	fmt.Fprintf(w, "Overall %d", len(data))
}
