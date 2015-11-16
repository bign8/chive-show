package sharder

import (
	"bytes"
	"sync"

	"appengine"
	"appengine/datastore"
)

// Reader creates a new shard reader to retrieve data from datastore
func Reader(c appengine.Context, name string) (*bytes.Buffer, error) {
	if name == "" {
		return nil, ErrInvalidName
	}

	var master shardMaster
	if err := datastore.Get(c, masterKey(c, name), &master); err != nil {
		panic(err)
		return nil, err
	}
	shards := (master.Size-1)/divisor + 1

	var wg sync.WaitGroup
	wg.Add(shards)
	data := make([]byte, master.Size)
	for i := 0; i < shards; i++ {
		go func(i int) {
			var shardData shard
			if err := datastore.Get(c, shardKey(c, name, i), &shardData); err != nil {
				panic(err)
			}
			// c.Infof("Out Data %d: %q", i, string(shardData.Data))

			end := i*divisor + divisor
			if end > master.Size {
				end = master.Size
			}
			copy(data[i*divisor:end], shardData.Data)
			wg.Done()
		}(i)
	}
	wg.Wait()

	return bytes.NewBuffer(data), nil
}
