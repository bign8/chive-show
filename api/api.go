package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bign8/chive-show/models"
)

// Init fired to initialize api
func Init(store models.Store) {
	http.Handle("/api/v1/post/random", handle(store.Random))
	http.Handle("/api/v1/post", handle(store.List))
	http.Handle("/api/v1/tags", tags(store.Tags))
}

func getURLCount(url *url.URL) int {
	val, err := strconv.Atoi(url.Query().Get("count"))
	if err != nil || val > 30 || val < 1 {
		return 2
	}
	return val
}

// Return a list of posts from the chive.
// Application always returns 200, use the payload to derive service/request failures.
func handle(fn func(context.Context, *models.ListOptions) (*models.ListResult, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		enc.SetEscapeHTML(false) // idk why I hate it so much, but I do!

		// Parse request parameters
		count := getURLCount(r.URL)
		log.Printf("INFO: Requested %v posts", count)

		// Fire the real request
		opts := &models.ListOptions{
			Count:  count,
			Cursor: r.URL.Query().Get("cursor"),
			Tag:    r.URL.Query().Get("tag"),
		}
		res, err := fn(r.Context(), opts)
		if err != nil {
			log.Printf("api(handle): %v", err)
			enc.Encode(response{
				Status: "error",
				Code:   http.StatusInternalServerError,
				Err:    "unable to fetch posts",
			})
			return
		}

		// fix up full response URLs based on incoming request
		r.URL.Host = r.Host
		if host := r.Header.Get("x-forwarded-host"); host != "" {
			r.URL.Host = host // work around test environment shortcomings
		}
		r.URL.Scheme = "http"
		if scheme := r.Header.Get("x-forwarded-proto"); scheme != "" {
			r.URL.Scheme = scheme
		}

		// Successful response!
		enc.Encode(response{
			Status: "success",
			Code:   http.StatusOK,
			Data:   res.Posts,
			Next:   toLink(r.URL, res.Next),
			Prev:   toLink(r.URL, res.Prev),
			Self:   toLink(r.URL, res.Self),
		})
	}
}

func tags(fn func(context.Context) (map[string]int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		enc.SetEscapeHTML(false) // idk why I hate it so much, but I do!

		// Fire the real request
		res, err := fn(r.Context())
		if err != nil {
			log.Printf("api(tags): %v", err)
			enc.Encode(response{
				Status: "error",
				Code:   http.StatusInternalServerError,
				Err:    "unable to fetch posts",
			})
			return
		}

		// Successful response!
		enc.Encode(response{
			Status: "success",
			Code:   http.StatusOK,
			Tags:   res,
		})
	}
}

type response struct {
	Status string         `json:"status"`
	Code   int            `json:"code"`
	Err    string         `json:"error,omitempty"`
	Data   []models.Post  `json:"data,omitempty"`
	Tags   map[string]int `json:"tags,omitempty"`
	Next   string         `json:"next_url,omitempty"`
	Prev   string         `json:"prev_url,omitempty"`
	Self   string         `json:"self_url,omitempty"`
}

func toLink(parent *url.URL, opts *models.ListOptions) string {
	if opts == nil || opts.Cursor == "" {
		return ``
	}
	vals := url.Values{
		"count":  {strconv.Itoa(opts.Count)},
		"cursor": {opts.Cursor},
	}
	if opts.Tag != "" {
		vals.Add("tag", opts.Tag)
	}
	return parent.ResolveReference(&url.URL{RawQuery: vals.Encode()}).String()
}
