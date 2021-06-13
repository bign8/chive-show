package cron

import (
	"context"
	"net/http"
	"net/http/httputil"

	"github.com/bign8/chive-show/models"
)

func CrawlHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bits, err := httputil.DumpRequest(r, true)
		if err != nil {
			panic(err)
		}
		w.Write(bits)
		// http.Error(w, "TODO", http.StatusNotImplemented)
	}
}

func fetchRssPage(ctx context.Context, link string) ([]models.Post, error) {
	return nil, nil
}
