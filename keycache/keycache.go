package keycache

import (
	"context"

	"google.golang.org/appengine/datastore"
)

const (
	// NAME is the table cache is stored in
	NAME = "DatastoreKeysCache"
)

func memcacheKey(name string) string {
	return NAME + ":" + name
}

func datastoreKey(c context.Context, name string) *datastore.Key {
	return datastore.NewKey(c, NAME, name, 0, nil)
}

// Object: entityKeys

// TODO: use prococal buffers (because why not) and have this be []byte
// TODO: or use array of strings with GobDecode / GobEncode (since that uses buffers already)
type entityKeys struct {
	Keys []entityKey
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
			StringID: key.StringID(),
			IntID:    key.IntID(),
		}
		if !duplicate[temp] {
			x.Keys = append(x.Keys, temp)
		}
	}
}

func (x *entityKeys) toKeys(c context.Context, name string) []*datastore.Key {
	keys := make([]*datastore.Key, len(x.Keys))
	for i, item := range x.Keys {
		keys[i] = item.toKey(c, name)
	}
	return keys
}

// Object: entityKey

// TODO: convert to using GobDecode / GobEncode
type entityKey struct {
	StringID string
	IntID    int64
}

func (x *entityKey) toKey(c context.Context, name string) *datastore.Key {
	return datastore.NewKey(c, name, x.StringID, x.IntID, nil)
}
