package keycache

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
)

func ResetKeys(c appengine.Context, name string) error {
  errc := make(chan error, 2)
  go func() {
    err := memcache.Delete(c, memcache_key(name))
    if err == memcache.ErrCacheMiss { err = nil }
    errc <- err
  }()
  go func() {
    err := datastore.Delete(c, datastore_key(c, name))
    if err == datastore.ErrNoSuchEntity { err = nil }
    errc <- err
  }()
  err1, err2 := <-errc, <-errc
  if err1 != nil {return err1}
  return err2
}
