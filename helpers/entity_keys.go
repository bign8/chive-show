package helpers

import (
  "encoding/json"

  "appengine"
  "appengine/datastore"
  "appengine/memcache"
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
  keyStorageKind  = "GeneralKeysOptimal"
)

func mc(name string) string {
  return keyStorageKind + ":" + name
}

func dbKeyToPostKey(k *datastore.Key) entityKey {
  return entityKey{
    StringID:  k.StringID(),
    IntID:     k.IntID(),
  }
}

func postKeyToDBKey(c appengine.Context, name string, k entityKey) *datastore.Key {
  return datastore.NewKey(c, name, k.StringID, k.IntID, nil)
}

// Count retrieves the value of the named counter.
func GetKeys(c appengine.Context, name string) ([]*datastore.Key, error) {
  var postKeys entityKeys

  // Check Memcache
  cacheItem, err := memcache.Get(c, mc(name))
  if err != nil && err != memcache.ErrCacheMiss {
    return nil, err
  }
  if err == nil {

    // Memcache HIT
    c.Infof("Memcache HIT")
    err = json.Unmarshal(cacheItem.Value, &postKeys)

  } else {

    // Memcache MISS
    c.Infof("Memcache MISS")
    key := datastore.NewKey(c, keyStorageKind, name, 0, nil) // Note: will need to be deleted until cron is updated
    err = datastore.Get(c, key, &postKeys)

    // Datastore MISS
    if err == datastore.ErrNoSuchEntity {
      c.Infof("Datastore MISS")
      err = nil
      keys, err := datastore.NewQuery(name).KeysOnly().GetAll(c, nil)
      if err != nil {
        return nil, err
      }
      postKeys.Keys = make([]entityKey, len(keys))
      for idx, item := range keys {
        postKeys.Keys[idx] = dbKeyToPostKey(item)
      }
      c.Infof("key %v", key)
      _, err = datastore.Put(c, key, &postKeys)
      c.Infof("err %v", err)
    }

    // Fork setting memcache so other things can run
    go func() {
      b, err := json.Marshal(postKeys)
      if err == nil {
        err = memcache.Set(c, &memcache.Item{
          Key:   mc(name),
          Value: b,
        })
      }
    }()
  }

  // Convert entityKey to real keys
  keys := make([]*datastore.Key, len(postKeys.Keys))
  for idx, item := range postKeys.Keys {
    keys[idx] = postKeyToDBKey(c, name, item)
  }
  return keys, err
}
