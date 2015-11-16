package sharder

import (
	"errors"
	"fmt"

	"appengine"
	"appengine/datastore"
)

const (
	masterKind = "shard-master"
	shardKind  = "shard-pieces"
	divisor    = 1e6 // 1MB
)

// ErrInvalidName because reasons
var ErrInvalidName = errors.New("Must provide name of sharded item")

func masterKey(c appengine.Context, name string) *datastore.Key {
	return datastore.NewKey(c, masterKind, name, 0, nil)
}

func shardKey(c appengine.Context, name string, idx int) *datastore.Key {
	return datastore.NewKey(c, shardKind, fmt.Sprintf("%s-%d", name, idx), 0, nil)
}

type shardMaster struct {
	Size int `datastore:"size"`
}

type shard struct {
	Data []byte
}
