package keycache

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
)

func AddKeys(c appengine.Context, name string, keys []*datastore.Key) error {
  var container entityKeys
  ds_key := datastore_key(c, name)
  mc_key := memcache_key(name)

  // Read
  _, err := memcache.Gob.Get(c, mc_key, &container)
  if err != nil {
    if err != memcache.ErrCacheMiss {
      return err
    }
    err = datastore.Get(c, ds_key, &container)
    if err != nil && err != datastore.ErrNoSuchEntity {
      return err
    }
  }

  // Update
  container.addKeys(keys)

  // Write
  errc := make(chan error, 2)
  go func() {
    _, err = datastore.Put(c, ds_key, &container)
    errc <- err
  }()
  go func() { // TODO: timeout if longer than Put
    errc <- memcache.Gob.Set(c, &memcache.Item{
      Key:    mc_key,
      Object: container,
    })
  }()
  err1, err2 := <-errc, <-errc
  if err1 != nil { return err1 }
  return err2
}
