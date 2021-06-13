package cron

import (
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/bign8/chive-show/appengine"
	"github.com/bign8/chive-show/models"
)

// var latestVersion = 0

func MigrateHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := log.New(io.MultiWriter(w, log.Default().Writer()), "migrate: ", 0)
		store, err := datastore.NewClient(r.Context(), appengine.ProjectID())
		if err != nil {
			info.Printf("Unable create datastore client: %v", err)
			return
		}
		var posts []models.Post
		keys, err := store.GetAll(r.Context(), datastore.NewQuery(`Post`), &posts)
		if err != nil {
			info.Printf("Unable to fetch all posts: %v", err)
			return
		}
		_, err = store.PutMulti(r.Context(), keys, posts)
		if err != nil {
			info.Printf("Unable to save %d posts: %v", len(posts), err)
			return
		}
		info.Printf("Migrated %d posts", len(posts))
	}
}
