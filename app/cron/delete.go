package cron

// TODO: delete this file once feed versioning is a thing

import (
  "app/models"
  "app/helpers/keycache"
  "appengine"
  "appengine/datastore"
  "fmt"
  "net/http"
)

func delete(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)

  // TODO: use helpers.keycache to help here
  q := datastore.NewQuery(models.DB_POST_TABLE).KeysOnly()
  keys, err := q.GetAll(c, nil)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // Batch Delete
  var del_keys []*datastore.Key
  for _, key := range keys {
    if del_keys == nil {
      del_keys = make([]*datastore.Key, 0)
    }
    del_keys = append(del_keys, key)
    if len(del_keys) > 50 {
      err = datastore.DeleteMulti(c, del_keys)
      c.Infof("Deleting Keys %v", del_keys)
      del_keys = nil
    }
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  if len(del_keys) > 0 {
    err = datastore.DeleteMulti(c, del_keys)
  }
  fmt.Fprintf(w, "%v\n%v\nDeleted", err, keycache.ResetKeys(c, models.DB_POST_TABLE))
}
