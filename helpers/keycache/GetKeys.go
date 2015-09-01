package keycache

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
  "time"
)

func GetKeys(c appengine.Context, name string) ([]*datastore.Key, error) {
  var container entityKeys

  // Check Memcache
  start := time.Now()
  _, err := memcache.Gob.Get(c, memcache_key(name), &container)
  c.Infof("Actual Memcache.Get: %s", time.Since(start))

  if err != nil {
    if err != memcache.ErrCacheMiss {
      return nil, err
    }

    key := datastore_key(c, name)
    start := time.Now()
    err = datastore.Get(c, key, &container)
    c.Infof("Actual Datastore.Get: %s", time.Since(start))

    // Datastore MISS
    if err == datastore.ErrNoSuchEntity {  // FYI: this is a costly operation
      c.Infof("Datastore MISS: Costly Query getting keys over \"%v\"", name)
      err = nil
      keys, err := datastore.NewQuery(name).KeysOnly().GetAll(c, nil)
      if err != nil {
        return nil, err
      }
      container.addKeys(keys)
      _, err = datastore.Put(c, key, &container)
    }

    // Fork setting memcache so other things can run
    go func() {
      memcache.Gob.Set(c, &memcache.Item{
        Key:    memcache_key(name),
        Object: container,
      })
    }()
  }

  // Convert entityKey to real keys
  keys := make([]*datastore.Key, len(container.Keys))
  for idx, item := range container.Keys {
    keys[idx] = datastore.NewKey(c, name, item.StringID, item.IntID, nil)
  }
  return keys, err
}
