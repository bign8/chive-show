package sharder

import (
	"sync"

	"appengine"
	"appengine/datastore"
)

// Writer shards and stores a byte String
func Writer(c appengine.Context, name string, data []byte) error {
	if name == "" {
		return ErrInvalidName
	}

	master := shardMaster{len(data)}
	shards := numShards(master.Size)

	// Store shardMaster
	if _, err := datastore.Put(c, masterKey(c, name), &master); err != nil {
		return err
	}

	// shard data and store shards
	var wg sync.WaitGroup
	wg.Add(shards)
	for i := 0; i < shards; i++ {
		go func(i int) {
			shardKey := shardKey(c, name, i)
			shardData := data[i*divisor:]
			if len(shardData) > divisor {
				shardData = data[:divisor]
			}
			s := shard{shardData}
			// w.ctx.Infof("Inn Data %d: %q", i, s.Data)
			if _, err := datastore.Put(c, shardKey, &s); err != nil {
				panic(err)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	return nil
}
