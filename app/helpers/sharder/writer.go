package sharder

import (
	"bytes"
	"errors"
	"time"

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
		key:  nil,
		buff: bytes.NewBufferString(""),
		name: name,
	}, nil
}

// Writer is the item that deals with writing sharded data
type Writer struct {
	buff *bytes.Buffer
	ctx  appengine.Context
	key  *ShardKey
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
	// TODO: datastore.RunInTransaction + go-routines with waitGroups

	length := w.buff.Len()
	shards := (length-1)/divisor + 1
	key := masterKey(w.ctx, w.name)

	// Store shardMaster
	master := shardMaster{
		Name:   w.name,
		Stamp:  time.Now(),
		Shards: shards,
		MD5:    "TO-IMPLEMENT",
		Size:   length,
	}
	if _, err := datastore.Put(w.ctx, key, &master); err != nil {
		panic(err)
		return err
	}

	// shard data and store shards
	data := w.buff.Bytes()
	for i := 0; i < shards; i++ {
		shardKey := shardKey(w.ctx, w.name, i)
		shardData := data[i*divisor:]
		if len(shardData) > divisor {
			shardData = data[:divisor]
		}
		s := shard{shardData}
		w.ctx.Infof("Inn Data %d: %q", i, s.Data)
		if _, err := datastore.Put(w.ctx, shardKey, &s); err != nil {
			panic(err)
			return err
		}
	}

	w.key = new(ShardKey)
	*w.key = ShardKey(key.Encode())
	w.buff = nil
	return nil
}

// Key returns the key of the sharded data.  Note: will return an error if not Closed.
func (w *Writer) Key() (*ShardKey, error) {
	if w.key == nil {
		return nil, errors.New("Writer must be closed before a Key is available")
	}
	return w.key, nil
}
