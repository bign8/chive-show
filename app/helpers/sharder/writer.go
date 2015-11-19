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

	// Attempt to get existing key
	key := masterKey(c, name)
	oldMaster := shardMaster{}
	oldShards := 0
	if datastore.Get(c, key, &oldMaster) == nil {
		oldShards = numShards(oldMaster.Size)
	}

	// Store shardMaster
	master := shardMaster{len(data)}
	shards := numShards(master.Size)
	if _, err := datastore.Put(c, key, &master); err != nil {
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

	// Delete shards that shouldn't be in datastore (write something smaller than before)
	if oldShards > shards {
		keys := make([]*datastore.Key, oldShards-shards)
		for i := shards; i < oldShards; i++ {
			keys[i-shards] = shardKey(c, name, i)
		}
		datastore.DeleteMulti(c, keys)
	}

	wg.Wait()
	return nil
}
