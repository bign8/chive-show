package crawler

import (
	"appengine"
	"appengine/datastore"
)

func Storage(c appengine.Context, in <-chan []Data) {
	runStorage(c, in)
}

type Store struct {
	XML []byte
}

func runStorage(c appengine.Context, in <-chan []Data) {
	var keys []*datastore.Key
	var items []Store
	for batch := range in {
		keys = make([]*datastore.Key, len(batch))
		items = make([]Store, len(batch))
		for i, item := range batch {
			keys[i] = datastore.NewKey(c, XML, item.KEY, 0, nil)
			items[i].XML = []byte(item.XML)
		}

		c.Infof("Storage: Storing %v", keys)
		_, err := datastore.PutMulti(c, keys, items)
		if err != nil {
			c.Errorf("Storage: Error storing batch %s", err)
		}
	}
}
