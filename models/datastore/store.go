package datastore

import (
	"context"
	"log"
	"math/rand"

	"cloud.google.com/go/datastore"
	"github.com/bign8/chive-show/keycache"
	"github.com/bign8/chive-show/models"
)

func NewStore(store *datastore.Client) (*Store, error) {
	// TODO: move datastore.Client here once CRON is consuming this
	return &Store{store: store}, nil
}

type Store struct {
	store *datastore.Client
}

var _ models.Store = (*Store)(nil)

func (s *Store) Random(ctx context.Context, opts *models.RandomOptions) (*models.RandomResult, error) {
	if opts == nil {
		panic("nil options == bad")
	}

	// Pull keys from post keys object
	keys, err := keycache.GetKeys(ctx, s.store, models.POST)
	if err != nil {
		log.Printf("ERR: keycache.GetKeys %v", err)
		return nil, err
	}
	if len(keys) < opts.Count {
		log.Printf("ERR: Not enough keys(%v) for count(%v)", len(keys), opts.Count)
		return nil, models.ErrNotEnough
	}

	// Randomize list of keys
	// TODO: remember seed and offset to create a NEXT link
	for i := range keys {
		j := rand.Intn(i + 1)
		keys[i], keys[j] = keys[j], keys[i]
	}

	// Pull posts from datastore
	data := make([]models.Post, opts.Count) // TODO: cache items in memcache too (make a helper)
	if err := s.store.GetMulti(ctx, keys[:opts.Count], data); err != nil {
		log.Printf("ERR: datastore.GetMulti %v", err)
		return nil, err
	}
	return &models.RandomResult{
		Posts: data,
		Next: &models.RandomOptions{
			Count: opts.Count,
			Cursor: "yeet",
			// TODO: the rest of the attributes
		},
	}, nil
}
