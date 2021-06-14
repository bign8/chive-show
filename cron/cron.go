package cron

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"

	"github.com/bign8/chive-show/models/datastore"
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

	http.Handle("/cron/rebuild", RebuildHandler(store)) // Step 0: rebuild from nothing (on project init)
	http.Handle("/cron/crawl", CrawlHandler(store))     // Step 1: search for posts (on cron schedule)
	http.Handle("/cron/mine", MineHandler(store))       // Step 2: load post metadata (from cron)
	http.Handle("/cron/migrate", MigrateHandler(store)) // Step 4: migrate stored data (on demand)
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

type errBadFetch struct {
	url  string
	code int
}

func (err errBadFetch) Error() string {
	return fmt.Sprintf(`got a %d when fetching %s`, err.code, err.url)
}

func fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, ErrFeedParse404
		}
		return nil, errBadFetch{url: url, code: resp.StatusCode}
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
