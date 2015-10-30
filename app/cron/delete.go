package cron

// TODO: delete this file once feed versioning is a thing

import (
	"fmt"
	"net/http"

	"github.com/bign8/chive-show/app/helpers/keycache"
	"github.com/bign8/chive-show/app/models"

	"appengine"
	"appengine/datastore"
)

func delete(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// TODO: use helpers.keycache to help here
	q := datastore.NewQuery(models.POST).KeysOnly()
	keys, err := q.GetAll(c, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Batch Delete
	var delKeys []*datastore.Key
	for _, key := range keys {
		if delKeys == nil {
			delKeys = make([]*datastore.Key, 0)
		}
		delKeys = append(delKeys, key)
		if len(delKeys) > 50 {
			err = datastore.DeleteMulti(c, delKeys)
			c.Infof("Deleting Keys %v", delKeys)
			delKeys = nil
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(delKeys) > 0 {
		err = datastore.DeleteMulti(c, delKeys)
	}
	fmt.Fprintf(w, "%v\n%v\nDeleted", err, keycache.ResetKeys(c, models.POST))
}
