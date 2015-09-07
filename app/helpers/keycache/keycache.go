package keycache

import (
  "appengine"
  "appengine/datastore"
)

const (
  DEFAULT_TIMEOUT  = 20 // TODO: implement This
  PERSISTENCE_NAME = "DatastoreKeysCache"
)

func memcache_key(name string) string {
  return PERSISTENCE_NAME + ":" + name
}

func datastore_key(c appengine.Context, name string) *datastore.Key {
  return datastore.NewKey(c, PERSISTENCE_NAME, name, 0, nil)
}

// Object: entityKeys

type entityKeys struct {
  Keys []entityKey  // TODO: use prococal buffers (because why not) and have this be []byte
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

func (x *entityKeys) toKeys(c appengine.Context, name string) []*datastore.Key {
  keys := make([]*datastore.Key, len(x.Keys))
  for i, item := range x.Keys {
    keys[i] = item.toKey(c, name)
  }
  return keys
}

// Object: entityKey

type entityKey struct {
  StringID string
  IntID    int64
}

func (x *entityKey) toKey(c appengine.Context, name string) *datastore.Key {
  return datastore.NewKey(c, name, x.StringID, x.IntID, nil)
}
