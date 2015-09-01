package keycache

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
)

func ResetKeys(c appengine.Context, name string) error {
  err := memcache.Delete(c, memcache_key(name))
  if err != nil && err != memcache.ErrCacheMiss {
    return err
  }
  return datastore.Delete(c, datastore_key(c, name))
}
