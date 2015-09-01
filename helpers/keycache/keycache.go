package keycache

import (
  "appengine"
  "appengine/datastore"
)

const (
  DEFAULT_TIMEOUT  = 20 // TODO: implement This
  PERSISTENCE_NAME = "DatastoreKeysCache"
)

type entityKeys struct {
  Keys []entityKey  // TODO: use prococal buffers (because why not) and have this be []byte
}

type entityKey struct {
  StringID string
  IntID    int64
}

func (x *entityKeys) addKeys(keys []*datastore.Key) {
  if x.Keys == nil {
    x.Keys = make([]entityKey, 0)
  }

  duplicate := make(map[entityKey]bool)
  for _, key := range x.Keys {
    duplicate[key] = true
  }

  for _, key := range keys {
    temp := entityKey{
      StringID:  key.StringID(),
      IntID:     key.IntID(),
    }
    if !duplicate[temp] {
      x.Keys = append(x.Keys, temp)
    }
  }
}

func memcache_key(name string) string {
  return PERSISTENCE_NAME + ":" + name
}

func datastore_key(c appengine.Context, name string) *datastore.Key {
  return datastore.NewKey(c, PERSISTENCE_NAME, name, 0, nil)
}
