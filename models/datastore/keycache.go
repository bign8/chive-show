package datastore

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/datastore"
)

func datastoreKey(name string) *datastore.Key {
	return datastore.NameKey(`DatastoreKeysCache`, name, nil)
}

// entityKeys stores entity keys using a gzip + json format for minimal work and okay compression
type entityKeys []entityKey
type entityKey struct {
	StringID string `json:"str,omitempty"`
	IntID    int64  `json:"int,omitempty"`
}

// Load Datastore PropertyLoadSaver Interface : https://pkg.go.dev/cloud.google.com/go/datastore#PropertyLoadSaver
func (keys *entityKeys) Load(c []datastore.Property) error {
	cmp, err := gzip.NewReader(bytes.NewReader(c[0].Value.([]byte)))
	if err != nil {
		return err
	}
	if err = json.NewDecoder(cmp).Decode(keys); err != nil {
		return err
	}
	return cmp.Close()
}

// Save Datastore PropertyLoadSaver Interface : https://pkg.go.dev/cloud.google.com/go/datastore#PropertyLoadSaver
func (keys entityKeys) Save() (props []datastore.Property, err error) {
	var buffer bytes.Buffer
	stream := gzip.NewWriter(&buffer)
	if err = json.NewEncoder(stream).Encode(&keys); err != nil {
		return nil, err
	}
	if err = stream.Close(); err != nil {
		return nil, err
	}
	return []datastore.Property{{
		Name:    "data",
		Value:   buffer.Bytes(),
		NoIndex: true,
	}}, nil
}

func (x *entityKeys) addKeys(keys []*datastore.Key) {
	duplicate := make(map[entityKey]bool)
	for _, key := range *x {
		duplicate[key] = true
	}
	for _, key := range keys {
		temp := entityKey{
			StringID: key.Name,
			IntID:    key.ID,
		}
		if !duplicate[temp] {
			*x = append(*x, temp)
		}
	}
}

func (x entityKeys) toKeys(name string) []*datastore.Key {
	keys := make([]*datastore.Key, len(x))
	for i, item := range x {
		if item.IntID != 0 {
			keys[i] = datastore.IDKey(name, item.IntID, nil)
		} else {
			keys[i] = datastore.NameKey(name, item.StringID, nil)
		}
	}
	return keys
}

// addKeys add keys to the context
func addKeys(c context.Context, store datastoreClient, name string, keys []*datastore.Key) error {
	start := time.Now()
	var container entityKeys
	ds := datastoreKey(name)
	err := store.Get(c, ds, &container)
	if err != nil && err != datastore.ErrNoSuchEntity {
		return err
	}
	before := len(container)
	container.addKeys(keys)
	_, err = store.Put(c, ds, &container)
	log.Printf("DEBU: addKeys took %s seconds to add %d keys", time.Since(start), len(container)-before) // TODO: use tracing spans
	return err
}

// GetKeys returns the keys for a particular item
func getKeys(c context.Context, store datastoreClient, name string) ([]*datastore.Key, error) {
	start := time.Now()
	var container entityKeys
	key := datastoreKey(name)
	err := store.Get(c, key, &container)

	// Datastore MISS
	if err == datastore.ErrNoSuchEntity { // FYI: this is a costly operation
		log.Printf("INFO: Datastore MISS: Costly Query getting keys over %q", name)
		err = nil
		keys, err := store.GetAll(c, datastore.NewQuery(name).KeysOnly(), nil)
		if err != nil {
			return nil, err
		}
		container.addKeys(keys)
		_, err = store.Put(c, key, &container)
		if err != nil {
			return nil, err
		}
	}
	keys := container.toKeys(name)
	log.Printf("DEBU: getKeys took %s seconds to get %d keys", time.Since(start), len(keys)) // TODO: use tracing spans
	return keys, nil
}

// // resetKeys resets all item keys
// func resetKeys(c context.Context, store *datastore.Client, name string) error {
// 	err := store.Delete(c, datastoreKey(name))
// 	if err == datastore.ErrNoSuchEntity {
// 		return nil
// 	}
// 	return err
// }
