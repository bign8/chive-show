package keycache

import (
	"context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

// ResetKeys resets all item keys
func ResetKeys(c context.Context, name string) error {
	errc := make(chan error, 2)
	go func() {
		err := memcache.Delete(c, memcacheKey(name))
		if err == memcache.ErrCacheMiss {
			err = nil
		}
		errc <- err
	}()
	go func() {
		err := datastore.Delete(c, datastoreKey(c, name))
		if err == datastore.ErrNoSuchEntity {
			err = nil
		}
		errc <- err
	}()
	err1, err2 := <-errc, <-errc
	if err1 != nil {
		return err1
	}
	return err2
}
