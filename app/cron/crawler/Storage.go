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
	case "vertex":
		store = runStorageVertex
	case "edge":
		store = runStorageEdge
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

// Puller pull items from datastore
// TODO: improve pulling performance (cache number of xml in stage_1, fan out pulling)
func Puller(c appengine.Context, loc string) <-chan string {
	out := make(chan string, 10000)

	go func() {
		defer close(out)
		q := datastore.NewQuery(loc)
		t := q.Run(c)
		for {
			var s Store
			_, err := t.Next(&s)
			if err == datastore.Done {
				break // No further entities match the query.
			}
			if err != nil {
				c.Errorf("fetching next Person: %v", err)
				break
			}

			// Do something with Person p and Key k
			out <- string(s.XML)
		}
	}()
	return out
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

func runStorageVertex(c appengine.Context, in <-chan []interface{}, x int, loc string) {
	var keys []*datastore.Key
	var items []Vertex

	for batch := range in {
		c.Infof("Storage %d: Storing Vertex chunk", x)
		keys = make([]*datastore.Key, len(batch))
		items = make([]Vertex, len(batch))
		for i, item := range batch {
			x := item.(Vertex)
			keys[i] = datastore.NewKey(c, loc, x.Type+":"+x.Value, 0, nil)
			items[i] = x
		}

		// c.Infof("Storage: Storing %v", keys)
		_, err := datastore.PutMulti(c, keys, items)
		if err != nil {
			c.Errorf("Storage %d: Error %s: %v %v", x, err, keys, items)
			panic(err)
		}
	}
}

func runStorageEdge(c appengine.Context, in <-chan []interface{}, x int, loc string) {
	var keys []*datastore.Key
	var items []Edge

	for batch := range in {
		c.Infof("Storage %d: Storing Edge chunk", x)
		keys = make([]*datastore.Key, len(batch))
		items = make([]Edge, len(batch))
		for i, item := range batch {
			x := item.(Edge)
			keys[i] = datastore.NewIncompleteKey(c, loc, nil)
			items[i] = x
		}

		// c.Infof("Storage: Storing %v", keys)
		_, err := datastore.PutMulti(c, keys, items)
		if err != nil {
			c.Errorf("Storage %d: Error %s: %v %v", x, err, keys, items)
			panic(err)
		}
	}
}
