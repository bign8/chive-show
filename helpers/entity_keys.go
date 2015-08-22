package helpers

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
  "time"
)

type entityKeys struct {
  Keys []entityKey
  // TODO: expires timestamp
}

type entityKey struct {
  StringID string
  IntID    int64
}

const (
  defaultTimeout   = 20
  keyStorageKind  = "DatastoreKeysCache"
)

func mcKey(name string) string {
  return keyStorageKind + ":" + name
}

// Count retrieves the value of the named counter.
func GetKeys(c appengine.Context, name string) ([]*datastore.Key, error) {
  var postKeys entityKeys

  // Check Memcache
  start := time.Now()
  _, err := memcache.Gob.Get(c, mcKey(name), &postKeys)
  c.Infof("Actual Memcache.Get: %s", time.Since(start))

  if err != nil {
    if err != memcache.ErrCacheMiss {
      return nil, err
    }

    key := datastore.NewKey(c, keyStorageKind, name, 0, nil)
    start := time.Now()
    err = datastore.Get(c, key, &postKeys)
    c.Infof("Actual Datastore.Get: %s", time.Since(start))

    // Datastore MISS
    if err == datastore.ErrNoSuchEntity {  // FYI: this is a costly operation
      c.Infof("Datastore MISS: Costly Query getting keys over \"%v\"", name)
      err = nil
      keys, err := datastore.NewQuery(name).KeysOnly().GetAll(c, nil)
      if err != nil {
        return nil, err
      }
      postKeys.Keys = make([]entityKey, len(keys))
      for idx, item := range keys {
        postKeys.Keys[idx] = entityKey{
          StringID:  item.StringID(),
          IntID:     item.IntID(),
        }
      }
      _, err = datastore.Put(c, key, &postKeys)
    }

    // Fork setting memcache so other things can run
    go func() {
      err = memcache.Gob.Set(c, &memcache.Item{
        Key:   mcKey(name),
        Object: postKeys,
      })
    }()
  }

  // Convert entityKey to real keys
  keys := make([]*datastore.Key, len(postKeys.Keys))
  for idx, item := range postKeys.Keys {
    keys[idx] = datastore.NewKey(c, name, item.StringID, item.IntID, nil)
  }
  return keys, err
}
