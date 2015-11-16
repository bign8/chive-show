package sharder

import (
	"bytes"
	"errors"
	"sync"

	"appengine"
	"appengine/datastore"
)

// NewWriter creates a new Sharder to write sharded data to datastore
func NewWriter(c appengine.Context, name string) (*Writer, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	return &Writer{
		ctx:  c,
		buff: bytes.NewBufferString(""),
		name: name,
	}, nil
}

// Writer is the item that deals with writing sharded data
type Writer struct {
	buff *bytes.Buffer
	ctx  appengine.Context
	name string
}

// Write pushed p bytes to underlying data stream.
func (w *Writer) Write(p []byte) (n int, err error) {
	if w.buff == nil {
		return 0, errors.New("Buffer is closed")
	}
	return w.buff.Write(p)
}

// Close finishes off the current buffer, shards and stores the data.
// Once Close is called, the user may call Key to get the key of the stored object.
func (w *Writer) Close() error {
	// TODO: datastore.RunInTransaction
	// TODO: delete existing shards greater than current

	length := w.buff.Len()
	shards := (length-1)/divisor + 1
	key := masterKey(w.ctx, w.name)

	// Store shardMaster
	master := shardMaster{
		Size: length,
	}
	if _, err := datastore.Put(w.ctx, key, &master); err != nil {
		panic(err)
		return err
	}

	// shard data and store shards
	data := w.buff.Bytes()
	var wg sync.WaitGroup
	wg.Add(shards)
	for i := 0; i < shards; i++ {
		go func(i int) {
			shardKey := shardKey(w.ctx, w.name, i)
			shardData := data[i*divisor:]
			if len(shardData) > divisor {
				shardData = data[:divisor]
			}
			s := shard{shardData}
			// w.ctx.Infof("Inn Data %d: %q", i, s.Data)
			if _, err := datastore.Put(w.ctx, shardKey, &s); err != nil {
				panic(err)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	w.buff = nil
	return nil
}
