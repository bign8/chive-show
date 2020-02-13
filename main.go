package main

import (
	"log"
	"net/http"

	"google.golang.org/appengine"

	"github.com/bign8/chive-show/api"
	"github.com/bign8/chive-show/cron"
)

func main() {
	http.HandleFunc("/", http.NotFound) // Default Handler
	http.HandleFunc("/_ah/warmup", func(w http.ResponseWriter, r *http.Request) {
		// Needed to be able to migrate traffic on promotion.
		log.Println("Warmup Done")
	})

	// Setup Other routes routes
	api.Init()
	cron.Init()

	appengine.Main()
}
