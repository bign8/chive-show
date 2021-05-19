package api

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"cloud.google.com/go/datastore"

	"github.com/bign8/chive-show/keycache"
	"github.com/bign8/chive-show/models"
)

// Init fired to initialize api
func Init(store *datastore.Client) {
	http.Handle("/api/v1/post/random", random(store))
}

// API Helper function
func getURLCount(url *url.URL) int {
	val, err := strconv.Atoi(url.Query().Get("count"))
	if err != nil || val > 30 || val < 1 {
		return 2
	}
	return val
}

// Actual API functions
func random(store *datastore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count := getURLCount(r.URL)
		log.Printf("INFO: Requested %v random posts", count)

		// Pull keys from post keys object
		keys, err := keycache.GetKeys(r.Context(), store, models.POST)
		if err != nil {
			log.Printf("ERR: keycache.GetKeys %v", err)
			fail(w, http.StatusInternalServerError, "keycache.GetKeys failure")
			return
		}

		// quick sanity check
		if len(keys) < count {
			log.Printf("ERR: Not enough keys(%v) for count(%v)", len(keys), count)
			fail(w, http.StatusInsufficientStorage, "keycache.GetKeys shortage")
			return
		}

		// Randomize list of keys
		for i := range keys {
			j := rand.Intn(i + 1)
			keys[i], keys[j] = keys[j], keys[i]
		}

		// Pull posts from datastore
		data := make([]models.Post, count) // TODO: cache items in memcache too (make a helper)
		if err := store.GetMulti(r.Context(), keys[:count], data); err != nil {
			log.Printf("ERR: datastore.GetMulti %v", err)
			fail(w, http.StatusInternalServerError, "datastore.GetMulti failure")
			return
		}

		pass(w, data, "")
	}
}

type core struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
}

type failure struct {
	core
	Err string `json:"error"`
}

type response struct {
	core
	Data []models.Post `json:"data"`
	Next string        `json:"next_url,omitempty"`
}

func fail(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	enc.Encode(failure{
		core: core{
			Status: "error",
			Code:   code,
		},
		Err: msg,
	})
}

func pass(w http.ResponseWriter, data []models.Post, next string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	enc.Encode(response{
		core: core{
			Status: "success",
			Code:   200,
		},
		Data: data,
		Next: next,
	})
}
