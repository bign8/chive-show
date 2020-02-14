package cron

// TODO: delete this file once feed versioning is a thing

import (
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/bign8/chive-show/keycache"
	"github.com/bign8/chive-show/models"
)

func delete(store *datastore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: use helpers.keycache to help here
		q := datastore.NewQuery(models.POST).KeysOnly()
		keys, err := store.GetAll(r.Context(), q, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Batch Delete
		// TODO(bign8): use offsets into keys rather than creating a new del keys array
		var delKeys []*datastore.Key
		for _, key := range keys {
			if delKeys == nil {
				delKeys = make([]*datastore.Key, 0)
			}
			delKeys = append(delKeys, key)
			if len(delKeys) > 50 {
				err = store.DeleteMulti(r.Context(), delKeys)
				log.Printf("Deleting Keys %v", delKeys)
				delKeys = nil
			}
			if err != nil {
				log.Printf("Delete Keys Error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if len(delKeys) > 0 {
			err = store.DeleteMulti(r.Context(), delKeys)
		}
		fmt.Fprintf(w, "%v\n%v\nDeleted", err, keycache.ResetKeys(r.Context(), store, models.POST))
	}
}
