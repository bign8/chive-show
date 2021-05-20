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
func random(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")

		// Parse request parameters
		count := getURLCount(r.URL)
		log.Printf("INFO: Requested %v random posts", count)

		posts, err := store.ListPosts(r.Context(), count)

		// Failed response
		if err != nil {
			log.Printf("api(store.ListPosts): %v", err)
			enc.Encode(failure{
				common: common{
					Status: "error",
					Code:   500,
				},
				Err: "unable to fetch posts",
			})
			return
		}

		// Successful response!
		enc.Encode(response{
			common: common{
				Status: "success",
				Code:   200,
			},
			Data: posts,
			// Next: next,
		})
	}
}

type common struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
}

type failure struct {
	common
	Err string `json:"error"`
}

type response struct {
	common
	Data []models.Post `json:"data"`
	Next string        `json:"next_url,omitempty"`
}
