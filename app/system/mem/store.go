package mem

import (
	"log"

	"github.com/bign8/chive-show/app/system"
)

type memStore struct {
	data map[system.Key]interface{}
}

func (store *memStore) Get(key *system.Key, dst interface{}) error {
	log.Printf("Get %s", key)
	obj, ok := store.data[*key]
	if !ok {
		return system.ErrDoesNotExist
	}
	dst = obj
	return nil
}

func (store *memStore) GetMulti(keys []*system.Key, dst []interface{}) error {
	log.Printf("GetMulti %s", keys)
	res := make([]interface{}, len(keys))
	for idx, key := range keys {
		if err := store.Get(key, res[idx]); err != nil {
			return err
		}
	}
	dst = res
	return nil
}

func (store *memStore) Put(key *system.Key, src interface{}) error {
	log.Printf("Put %s", key)
	store.data[*key] = src
	return nil
}

func (store *memStore) PutMulti(keys []*system.Key, src []interface{}) error {
	log.Printf("PutMulti %s", keys)
	for idx, key := range keys {
		if err := store.Put(key, src[idx]); err != nil {
			return err
		}
	}
	return nil
}

func (store *memStore) Delete(key *system.Key) error {
	log.Printf("Delete %s", key)
	delete(store.data, *key)
	return nil
}

func (store *memStore) DeleteMulti(keys []*system.Key) error {
	log.Printf("DeleteMulti %s", keys)
	for _, key := range keys {
		if err := store.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func (store *memStore) Query() error {
	return nil
}
