package datastore

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/datastore"

	"github.com/bign8/chive-show/models"
)

func TestRandom(t *testing.T) {
	s := &Store{
		getKeys: func(c context.Context, store datastoreClient, name string) ([]*datastore.Key, error) {
			return []*datastore.Key{
				datastore.IDKey(models.POST, 1, nil),
				datastore.IDKey(models.POST, 2, nil),
				datastore.IDKey(models.POST, 3, nil),
				datastore.IDKey(models.POST, 4, nil),
				datastore.IDKey(models.POST, 5, nil),
			}, nil
		},
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
	res, err := s.Random(context.TODO(), &models.ListOptions{
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
	getAll   func(interface{}) ([]*datastore.Key, error)
}

func (f *fake) Run(context.Context, *datastore.Query) *datastore.Iterator { return nil }
func (f *fake) Get(context.Context, *datastore.Key, interface{}) error    { return errors.New("TODO") }
func (f *fake) GetAll(_ context.Context, q *datastore.Query, obj interface{}) ([]*datastore.Key, error) {
	if f.getAll == nil {
		return nil, nil
	}
	return f.getAll(obj)
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
func (f *fake) RunInTransaction(context.Context, func(*datastore.Transaction) error, ...datastore.TransactionOption) (*datastore.Commit, error) {
	return nil, nil // can't make a transaction w/o internal access to datastore package
}
