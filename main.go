package main

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/bign8/chive-show/api"
	"github.com/bign8/chive-show/appengine"
	"github.com/bign8/chive-show/cron"
)

func main() {
	// http.HandleFunc("/", http.NotFound) // Default Handler
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/_ah/warmup", func(w http.ResponseWriter, r *http.Request) {
		// Needed to be able to migrate traffic on promotion.
		log.Println("Warmup Done")
		http.Error(w, "warm!", http.StatusOK)
	})

	// https://cloud.google.com/docs/authentication/production
	// GOOGLE_APPLICATION_CREDENTIALS=<path-to>/service-account.json
	store, err := datastore.NewClient(context.Background(), "crucial-alpha-706")
	if err != nil {
		panic(err)
	}

	// Setup Other routes routes
	api.Init(store)
	cron.Init(store)

	appengine.Main()
}
