package main

import (
	"log"
	"net/http"

	"github.com/bign8/chive-show/api"
	"github.com/bign8/chive-show/appengine"
	"github.com/bign8/chive-show/cron"
	"github.com/bign8/chive-show/models/datastore"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/_ah/warmup", func(w http.ResponseWriter, r *http.Request) {
		// Needed to be able to migrate traffic on promotion.
		log.Println("Warmup Done")
		http.Error(w, "warm!", http.StatusOK)
	})

	// Setup routes that require storage access
	store, err := datastore.NewStore()
	if err != nil {
		panic(err)
	}
	api.Init(store)
	cron.Init(store)

	appengine.Main()
}
