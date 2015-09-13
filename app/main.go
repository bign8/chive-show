package app

import (
  "github.com/bign8/chive-show/app/api"
  "github.com/bign8/chive-show/app/cron"
  "net/http"
)

func init() {
  http.HandleFunc("/", http.NotFound)  // Default Handler

  // Setup Other routes routes
  api.Init()
  cron.Init()
}
