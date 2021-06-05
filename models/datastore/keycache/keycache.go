package keycache

// TODO: move this package into models/datastore once cron is updated

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/datastore"
)

// TODO: have all these APIs NOT require NAME!!! it's literally in the keys :cry:

const (
	// KIND is the table cache is stored in
	KIND = "DatastoreKeysCache"
)

func memcacheKey(name string) string {
	return KIND + ":" + name
}

func datastoreKey(name string) *datastore.Key {
	return datastore.NameKey(KIND, name, nil)
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
			StringID: key.Name,
			IntID:    key.ID,
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

func (x *entityKey) toKey(c context.Context, kind string) *datastore.Key {
	if x.IntID != 0 {
		return datastore.IDKey(kind, x.IntID, nil)
	}
	return datastore.NameKey(kind, x.StringID, nil)
}

// AddKeys add keys to the context
func AddKeys(c context.Context, store DatastoreClient, name string, keys []*datastore.Key) error {
	var container entityKeys
	ds := datastoreKey(name)
	// mc := memcacheKey(name)

	// // Read
	// _, err := memcache.Gob.Get(c, mc, &container)
	// if err != nil {
	// 	if err != memcache.ErrCacheMiss {
	// 		return err
	// 	}
	err := store.Get(c, ds, &container)
	if err != nil && err != datastore.ErrNoSuchEntity {
		return err
	}
	// }

	// Update
	container.addKeys(keys)

	// Write
	// errc := make(chan error, 2)
	// go func() {
	_, err = store.Put(c, ds, &container)
	return err
	// 	errc <- err
	// }()
	// go func() { // TODO: timeout if longer than Put
	// 	errc <- memcache.Gob.Set(c, &memcache.Item{
	// 		Key:    mc,
	// 		Object: container,
	// 	})
	// }()
	// err1, err2 := <-errc, <-errc
	// if err1 != nil {
	// 	return err1
	// }
	// return err2
}

type DatastoreClient interface {
	Get(context.Context, *datastore.Key, interface{}) error
	GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error)
	Put(context.Context, *datastore.Key, interface{}) (*datastore.Key, error)
}

// GetKeys returns the keys for a particular item
func GetKeys(c context.Context, store DatastoreClient, name string) ([]*datastore.Key, error) {
	var container entityKeys

	// // Check Memcache
	// start := time.Now()
	// _, err := memcache.Gob.Get(c, memcacheKey(name), &container)
	// log.Printf("INFO: Actual Memcache.Get: %s", time.Since(start))
	//
	// if err != nil {
	// 	if err != memcache.ErrCacheMiss {
	// 		return nil, err
	// 	}

	key := datastoreKey(name)
	start := time.Now()
	err := store.Get(c, key, &container)
	log.Printf("INFO: Actual Datastore.Get: %s", time.Since(start))

	// Datastore MISS
	if err == datastore.ErrNoSuchEntity { // FYI: this is a costly operation
		log.Printf("INFO: Datastore MISS: Costly Query getting keys over \"%v\"", name)
		err = nil
		keys, err := store.GetAll(c, datastore.NewQuery(name).KeysOnly(), nil)
		if err != nil {
			return nil, err
		}
		container.addKeys(keys)
		_, err = store.Put(c, key, &container)
	}

	// // Fork setting memcache so other things can run
	// done := make(chan error, 1)
	// go func() {
	// 	done <- memcache.Gob.Set(c, &memcache.Item{
	// 		Key:    memcacheKey(name),
	// 		Object: container,
	// 	})
	// }()
	// select {
	// case err = <-done:
	// case <-time.After(3 * time.Millisecond):
	// }
	// }
	return container.toKeys(c, name), err
}

// ResetKeys resets all item keys
func ResetKeys(c context.Context, store *datastore.Client, name string) error {
	// errc := make(chan error, 2)
	// go func() {
	// 	err := memcache.Delete(c, memcacheKey(name))
	// 	if err == memcache.ErrCacheMiss {
	// 		err = nil
	// 	}
	// 	errc <- err
	// }()
	// go func() {
	err := store.Delete(c, datastoreKey(name))
	if err == datastore.ErrNoSuchEntity {
		err = nil
	}
	return err
	// 	errc <- err
	// }()
	// err1, err2 := <-errc, <-errc
	// if err1 != nil {
	// 	return err1
	// }
	// return err2
}
