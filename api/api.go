package api

import (
	"encoding/json"
	"fmt"
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		count := getURLCount(r.URL)
		log.Printf("INFO: Requested %v random posts", count)
		result := NewJSONResponse(http.StatusInternalServerError, "Unknown Error", nil)

		// Pull keys from post keys object
		keys, err := keycache.GetKeys(r.Context(), store, models.POST)
		if err != nil {
			log.Printf("ERR: heleprs.GetKeys %v", err)
			result.Msg = "Error with keycache GetKeys"
		} else if len(keys) < count {
			log.Printf("ERR: Not enough keys(%v) for count(%v)", len(keys), count)
			result.Msg = "Basically empty datastore"
		} else {

			// Randomize list of keys
			for i := range keys {
				j := rand.Intn(i + 1)
				keys[i], keys[j] = keys[j], keys[i]
			}

			// Pull posts from datastore
			data := make([]models.Post, count) // TODO: cache items in memcache too (make a helper)
			if err := store.GetMulti(r.Context(), keys[:count], data); err != nil {
				log.Printf("ERR: datastore.GetMulti %v", err)
				result.Msg = "Error with datastore GetMulti"
			} else {
				result = NewJSONResponse(http.StatusOK, "Your amazing data awaits", data)
			}
		}

		if err = result.write(w); err != nil {
			log.Printf("ERR: result.write %v", err)
		}
	}
}

var successCodes = map[int]bool{
	200: true, // 200 OK
}

// JSONResponse a wrapper for default responses
type JSONResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
}

// NewJSONResponse generate a JSONResponse
func NewJSONResponse(code int, message string, data interface{}) JSONResponse {
	status := "error"
	if successCodes[code] {
		status = "success"
	}
	return JSONResponse{
		Code:   code,
		Data:   data,
		Msg:    message,
		Status: status,
	}
}

func (res *JSONResponse) write(w http.ResponseWriter) error {
	items, err := json.MarshalIndent(&res, "", "  ")
	var out string
	if err != nil {
		out = "{\"status\":\"error\",\"code\":500,\"data\":null,\"msg\":\"Error marshaling data\"}"
	} else {
		out = string(items)
	}
	fmt.Fprint(w, out)
	return err
}
