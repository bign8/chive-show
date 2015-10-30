package api

import (
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bign8/chive-show/app/helpers/keycache"
	"github.com/bign8/chive-show/app/models"
	"gopkg.in/mjibson/v1/appstats"

	"appengine"
	"appengine/datastore"
)

func Init() {
	http.Handle("/api/v1/post/random", appstats.NewHandler(random))
}

// API Helper function
func get_url_count(url *url.URL) int {
	val, err := strconv.Atoi(url.Query().Get("count"))
	if err != nil || val > 30 || val < 1 {
		return 2
	}
	return val
}

// Actual API functions
func random(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	count := get_url_count(r.URL)
	c.Infof("Requested %v random posts", count)
	result := NewJsonResponse(500, "Unknown Error", nil)

	// Pull keys from post keys object
	keys, err := keycache.GetKeys(c, models.DB_POST_TABLE)
	if err != nil {
		c.Errorf("heleprs.GetKeys %v", err)
		result.Msg = "Error with keycache GetKeys"
	} else if len(keys) < count {
		c.Errorf("Not enough keys(%v) for count(%v)", len(keys), count)
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
			c.Errorf("datastore.GetMulti %v", err)
			result.Msg = "Error with datastore GetMulti"
		} else {
			result = NewJsonResponse(200, "Your amazing data awaits", data)
		}
	}

	if err = result.write(w); err != nil {
		c.Errorf("result.write %v", err)
	}
}
