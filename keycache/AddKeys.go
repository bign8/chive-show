package keycache

import (
	"context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

// AddKeys add keys to the context
func AddKeys(c context.Context, name string, keys []*datastore.Key) error {
	var container entityKeys
	ds := datastoreKey(c, name)
	mc := memcacheKey(name)

	// Read
	_, err := memcache.Gob.Get(c, mc, &container)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			return err
		}
		err = datastore.Get(c, ds, &container)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
	}

	// Update
	container.addKeys(keys)

	// Write
	errc := make(chan error, 2)
	go func() {
		_, err = datastore.Put(c, ds, &container)
		errc <- err
	}()
	go func() { // TODO: timeout if longer than Put
		errc <- memcache.Gob.Set(c, &memcache.Item{
			Key:    mc,
			Object: container,
		})
	}()
	err1, err2 := <-errc, <-errc
	if err1 != nil {
		return err1
	}
	return err2
}
