package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bign8/chive-show/models"
)

// Init fired to initialize api
func Init(store models.Store) {
	http.Handle("/api/v1/post/random", random(store))
}

func getURLCount(url *url.URL) int {
	val, err := strconv.Atoi(url.Query().Get("count"))
	if err != nil || val > 30 || val < 1 {
		return 2
	}
	return val
}

// Return a list of random posts from the chive.
// Application always returns 200, use the payload to derive service/request failures.
func random(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		enc.SetEscapeHTML(false) // idk why I hate it so much, but I do!

		// Parse request parameters
		count := getURLCount(r.URL)
		log.Printf("INFO: Requested %v random posts", count)

		// Fire the real request
		opts := &models.RandomOptions{
			Count:  count,
			Cursor: r.URL.Query().Get("cursor"),
		}
		res, err := store.Random(r.Context(), opts)
		if err != nil {
			log.Printf("api(store.Random): %v", err)
			enc.Encode(response{
				Status: "error",
				Code:   http.StatusInternalServerError,
				Err:    "unable to fetch posts",
			})
			return
		}

		// fix up full response URLs based on incoming request
		r.URL.Host = r.Header.Get("x-forwarded-host")
		r.URL.Scheme = r.Header.Get("x-forwarded-proto")

		// Successful response!
		enc.Encode(response{
			Status: "success",
			Code:   http.StatusOK,
			Data:   res.Posts,
			Next:   toLink(r.URL, res.Next),
			Prev:   toLink(r.URL, res.Prev),
		})
	}
}

type response struct {
	Status string        `json:"status"`
	Code   int           `json:"code"`
	Err    string        `json:"error,omitempty"`
	Data   []models.Post `json:"data,omitempty"`
	Next   string        `json:"next_url,omitempty"`
	Prev   string        `json:"prev_url,omitempty"`
}

func toLink(parent *url.URL, opts *models.RandomOptions) string {
	if opts == nil || opts.Cursor == "" {
		return ``
	}
	vals := url.Values{
		"count":  {strconv.Itoa(opts.Count)},
		"cursor": {opts.Cursor},
	}
	return parent.ResolveReference(&url.URL{RawQuery: vals.Encode()}).String()
}
