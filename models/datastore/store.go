package datastore

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

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

	var (
		seed     int64
		capacity int64
		offset   int
		err      error
	)
	if opts.Cursor != "" {
		parts := strings.SplitN(opts.Cursor, "~", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid cursor, expected 3 parts, got %d", len(parts))
		}
		offset, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid cursor(0): %v", err)
		}
		seed, err = strconv.ParseInt(parts[1], 36, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor(1): %v", err)
		}
		capacity, err = strconv.ParseInt(parts[2], 36, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor(2): %v", err)
		}
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

	// TODO: filter keys to be only those included in the first scrape (find max maybe?)
	// WARNING: theChive does not necessarily publish posts in their creation order, so this is BRITTLE!
	if capacity == 0 {
		for _, key := range keys {
			if key.ID > capacity {
				capacity = key.ID
			}
		}
	} else {
		// TODO: delete keys that are newer than "capacity"
		log.Printf("WIP!")
	}

	// Initialize seed to random seed if none provided
	if seed == 0 {
		seed = rand.Int63()
	}

	// Randomize list of keys
	rnd := rand.New(rand.NewSource(seed))
	for i := range keys {
		j := rnd.Intn(i + 1)
		keys[i], keys[j] = keys[j], keys[i]
	}

	// Pull posts from datastore
	// TODO: include limit logic here
	data := make([]models.Post, opts.Count) // TODO: cache items in memcache too (make a helper)
	if err := s.store.GetMulti(ctx, keys[:opts.Count], data); err != nil {
		log.Printf("ERR: datastore.GetMulti %v", err)
		return nil, err
	}
	return &models.RandomResult{
		Posts: data,
		Next: &models.RandomOptions{
			Count:  opts.Count,
			Cursor: strconv.Itoa(offset+opts.Count) + "~" + strconv.FormatInt(seed, 36) + "~" + strconv.FormatInt(capacity, 36),
			// TODO: remove if no next link is possible
		},
		Prev: &models.RandomOptions{
			Count:  opts.Count,
			Cursor: strconv.Itoa(offset-opts.Count) + "~" + strconv.FormatInt(seed, 36) + "~" + strconv.FormatInt(capacity, 36),
			// TODO: remove if no previous link possible
		},
	}, nil
}
