package api

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bign8/chive-show/app/helpers/keycache"
	"github.com/bign8/chive-show/app/models"
	"github.com/mjibson/appstats"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// Init fired to initialize api
func Init() {
	http.Handle("/api/v1/post/random", appstats.NewHandler(random))
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
func random(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	count := getURLCount(r.URL)
	log.Infof(c, "Requested %v random posts", count)
	result := NewJSONResponse(http.StatusInternalServerError, "Unknown Error", nil)

	// Pull keys from post keys object
	keys, err := keycache.GetKeys(c, models.POST)
	if err != nil {
		log.Errorf(c, "heleprs.GetKeys %v", err)
		result.Msg = "Error with keycache GetKeys"
	} else if len(keys) < count {
		log.Errorf(c, "Not enough keys(%v) for count(%v)", len(keys), count)
		result.Msg = "Basically empty datastore"
	} else {

		// Randomize list of keys
		for i := range keys {
			j := rand.Intn(i + 1)
			keys[i], keys[j] = keys[j], keys[i]
		}

		// Pull posts from datastore
		data := make([]models.Post, count) // TODO: cache items in memcache too (make a helper)
		if err := datastore.GetMulti(c, keys[:count], data); err != nil {
			log.Errorf(c, "datastore.GetMulti %v", err)
			result.Msg = "Error with datastore GetMulti"
		} else {
			result = NewJSONResponse(http.StatusOK, "Your amazing data awaits", data)
		}
	}

	if err = result.write(w); err != nil {
		log.Errorf(c, "result.write %v", err)
	}
}
