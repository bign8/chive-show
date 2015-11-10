package crawler

import (
	"sync"

	"appengine"
	"appengine/datastore"
)

func Storage(c appengine.Context, in <-chan []Data, workers int) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		go func(x int) {
			runStorage(c, in, x)
			wg.Done()
		}(i)
	}
	wg.Add(workers)
	wg.Wait()
}

type Store struct {
	XML []byte
}

func runStorage(c appengine.Context, in <-chan []Data, x int) {
	var keys []*datastore.Key
	var items []Store
	for batch := range in {
		c.Infof("Storage %d: Storing chunk", x)
		keys = make([]*datastore.Key, len(batch))
		items = make([]Store, len(batch))
		for i, item := range batch {
			keys[i] = datastore.NewKey(c, XML, item.KEY, 0, nil)
			items[i].XML = []byte(item.XML)
		}

		// c.Infof("Storage: Storing %v", keys)
		_, err := datastore.PutMulti(c, keys, items)
		if err != nil {
			c.Errorf("Storage: Error storing batch %s", err)
		}
	}
}
