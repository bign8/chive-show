package main

import "net/http"

// This is where Google App Engine sets up which handler lives at the root url.
func init() {
  // Immediately enter the main app.
  main()
}

func main() {
  http.HandleFunc("/", http.NotFound)  // Default Handler too

  // Setup Other routes routes
  api()
  cron()
}
