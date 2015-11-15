package crawler

import (
	"sync"

	"appengine"
	"appengine/datastore"
)

// Storage push items to datastore
func Storage(c appengine.Context, in <-chan []interface{}, workers int, loc string) {
	var store func(c appengine.Context, in <-chan []interface{}, x int, loc string)

	switch loc {
	case XML:
		store = runStorageData
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		go func(x int) {
			store(c, in, x, loc)
			wg.Done()
		}(i)
	}
	wg.Add(workers)
	wg.Wait()
}

// Store single xml item to put in storage
type Store struct {
	XML []byte
}

func runStorageData(c appengine.Context, in <-chan []interface{}, x int, loc string) {
	var keys []*datastore.Key
	var items []Store

	for batch := range in {
		c.Infof("Storage %d: Storing Post chunk", x)
		keys = make([]*datastore.Key, len(batch))
		items = make([]Store, len(batch))
		for i, item := range batch {
			x := item.(Data)
			keys[i] = datastore.NewKey(c, loc, x.KEY, 0, nil)
			items[i] = Store{[]byte(x.XML)}
		}

		// c.Infof("Storage: Storing %v", keys)
		_, err := datastore.PutMulti(c, keys, items)
		if err != nil {
			c.Errorf("Storage %d: Error %s: %v %v", x, err, keys, items)
			panic(err)
		}
	}
}
