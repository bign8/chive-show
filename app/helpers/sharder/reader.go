package sharder

import (
	"bytes"

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

	data := make([]byte, master.Size)
	for i := 0; i < master.Shards; i++ {
		var shardData shard
		if err := datastore.Get(c, shardKey(c, name, i), &shardData); err != nil {
			return nil, err
		}
		c.Infof("Out Data %d: %q", i, string(shardData.Data))

		end := i*divisor + divisor
		if end > master.Size {
			end = master.Size
		}
		copy(data[i*divisor:end], shardData.Data)
	}

	return bytes.NewBuffer(data), nil
}
