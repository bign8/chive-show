package datastore

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/datastore"

	"github.com/bign8/chive-show/keycache"
	"github.com/bign8/chive-show/models"
)

func TestRandom(t *testing.T) {
	getKeys = func(c context.Context, store keycache.DatastoreClient, name string) ([]*datastore.Key, error) {
		return []*datastore.Key{
			datastore.IDKey(models.POST, 1, nil),
			datastore.IDKey(models.POST, 2, nil),
			datastore.IDKey(models.POST, 3, nil),
			datastore.IDKey(models.POST, 4, nil),
			datastore.IDKey(models.POST, 5, nil),
		}, nil
	}
	s := &Store{
		store: &fake{
			getMulti: func(keys []*datastore.Key, obj interface{}) error {
				if len(keys) != 3 {
					t.Errorf("Expected getMulti with 3 keys, got %d", len(keys))
				}
				list := obj.([]models.Post)
				for i, post := range list {
					post.ID = int64(i)
					list[i] = post
				}
				return nil
			},
		},
	}
	res, err := s.Random(context.TODO(), &models.RandomOptions{
		Count:  3,
		Cursor: "0~2~0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Posts) != 3 {
		t.Errorf("Unexpected number of posts: %d", len(res.Posts))
	}
	if res.Next.Cursor != "3~2~5" {
		t.Errorf("Unexpected next cursor: %q", res.Next.Cursor)
	}
}

type fake struct {
	getMulti func([]*datastore.Key, interface{}) error
}

func (f *fake) Get(context.Context, *datastore.Key, interface{}) error { return errors.New("TODO") }
func (f *fake) GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error) {
	return nil, errors.New("TODO")
}
func (f *fake) Put(context.Context, *datastore.Key, interface{}) (*datastore.Key, error) {
	return nil, errors.New("TODO")
}
func (f *fake) PutMulti(context.Context, []*datastore.Key, interface{}) ([]*datastore.Key, error) {
	return nil, errors.New("TODO")
}
func (f *fake) GetMulti(_ context.Context, keys []*datastore.Key, obj interface{}) error {
	return f.getMulti(keys, obj)
}
