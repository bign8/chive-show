package datastore

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/googleapis/google-cloud-go-testing/datastore/dsiface"
)

type fakeClient struct {
	dsiface.Client
	getMulti func([]*datastore.Key, interface{}) error
	getAll   func(interface{}) ([]*datastore.Key, error)
	txn      dsiface.Transaction
}

func (f *fakeClient) GetAll(_ context.Context, q *datastore.Query, obj interface{}) ([]*datastore.Key, error) {
	if f.getAll == nil {
		return nil, nil
	}
	return f.getAll(obj)
}

func (f *fakeClient) GetMulti(_ context.Context, keys []*datastore.Key, obj interface{}) error {
	return f.getMulti(keys, obj)
}

func (f *fakeClient) RunInTransaction(_ context.Context, fn func(dsiface.Transaction) error, opts ...datastore.TransactionOption) (dsiface.Commit, error) {
	return nil, fn(f.txn)
}

type fakeTransaction struct {
	dsiface.Transaction
}

func (tx *fakeTransaction) GetMulti(keys []*datastore.Key, obj interface{}) error {
	return nil
}

func (tx *fakeTransaction) PutMulti(keys []*datastore.Key, obj interface{}) ([]*datastore.PendingKey, error) {
	return nil, nil
}
