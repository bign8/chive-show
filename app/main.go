package app

import (
  "app/api"
  "app/cron"
  "net/http"
)

func init() {
  http.HandleFunc("/", http.NotFound)  // Default Handler

  // Setup Other routes routes
  api.Init()
  cron.Init()
}
