package sharder

import (
	"errors"
	"fmt"
	"time"

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

// ShardKey is an identifying string for shards
type ShardKey string

func (sk *ShardKey) String() string {
	return fmt.Sprint(*sk)
}

// newKey takes the name of a file and creates a ShardKey
func newKey(c appengine.Context, name string) ShardKey {
	return ShardKey(masterKey(c, name).Encode())
}

func masterKey(c appengine.Context, name string) *datastore.Key {
	return datastore.NewKey(c, masterKind, name, 0, nil)
}

func shardKey(c appengine.Context, name string, idx int) *datastore.Key {
	return datastore.NewKey(c, shardKind, fmt.Sprintf("%s-%d", name, idx), 0, nil)
}

// ShardInfo implements the io.writer interface and allows for sharding data
type ShardInfo struct {
	Key          ShardKey
	CreationTime time.Time
	Size         int
	MD5          string
}

type shardMaster struct {
	Name   string    `datastore:"name"`
	Stamp  time.Time `datastore:"stamp"`
	Shards int       `datastore:"shards"`
	MD5    string    `datastore:"md5_hash"`
	Size   int       `datastore:"size"`
}

func (sm *shardMaster) toInfo(c appengine.Context) *ShardInfo {
	return &ShardInfo{
		Key:          newKey(c, sm.Name),
		CreationTime: sm.Stamp,
		Size:         sm.Size,
		MD5:          sm.MD5,
	}
}

type shard struct {
	Data []byte
}
