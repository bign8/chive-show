package main

import (
	"net/http"

	"google.golang.org/appengine"

	"github.com/bign8/chive-show/api"
	"github.com/bign8/chive-show/cron"
)

func main() {
	http.HandleFunc("/", http.NotFound) // Default Handler

	// Setup Other routes routes
	api.Init()
	cron.Init()

	appengine.Main()
}
