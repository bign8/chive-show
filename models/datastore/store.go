package datastore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"go.opencensus.io/trace"
	"google.golang.org/api/iterator"

	"github.com/bign8/chive-show/appengine"
	"github.com/bign8/chive-show/models"
)

func NewStore() (*Store, error) {
	// https://cloud.google.com/docs/authentication/production
	// GOOGLE_APPLICATION_CREDENTIALS=<path-to>/service-account.json
	store, err := datastore.NewClient(context.Background(), appengine.ProjectID())
	if err != nil {
		return nil, err
	}
	if os.Getenv("REBUILD") != "" {
		rebuildTags(store)
	}
	return &Store{
		store:   store,
		getKeys: getKeys,
	}, nil
}

type Store struct {
	store   datastoreClient
	stash   map[int64]bool
	getKeys func(c context.Context, store datastoreClient, name string) ([]*datastore.Key, error) // for testing
}

// allow faking out of the datastore for unit tests
type datastoreClient interface {
	Get(context.Context, *datastore.Key, interface{}) error
	GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error)
	Put(context.Context, *datastore.Key, interface{}) (*datastore.Key, error)
	GetMulti(context.Context, []*datastore.Key, interface{}) error
	PutMulti(context.Context, []*datastore.Key, interface{}) ([]*datastore.Key, error)
	Run(context.Context, *datastore.Query) *datastore.Iterator
	RunInTransaction(ctx context.Context, f func(tx *datastore.Transaction) error, opts ...datastore.TransactionOption) (cmt *datastore.Commit, err error)
}

var _ models.Store = (*Store)(nil)

func (s *Store) Random(rctx context.Context, opts *models.ListOptions) (*models.ListResult, error) {
	ctx, span := trace.StartSpan(rctx, "store.Random")
	defer span.End()
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
	keys, err := s.getKeys(ctx, s.store, models.POST)
	if err != nil {
		log.Printf("ERR: getKeys %v", err)
		return nil, err
	}
	if len(keys) < opts.Count || len(keys) < offset {
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
	max := opts.Count + offset
	if max > len(keys) {
		max = len(keys)
	}
	data := make([]models.Post, opts.Count) // TODO: cache items in memcache too (make a helper)
	if err := s.store.GetMulti(ctx, keys[offset:max], data); err != nil {
		log.Printf("ERR: datastore.GetMulti %v", err)
		return nil, err
	}

	// Setup cursors for next go around
	var (
		next *models.ListOptions
		prev *models.ListOptions
	)
	if max != len(keys) {
		next = &models.ListOptions{
			Count:  opts.Count,
			Cursor: strconv.Itoa(max) + "~" + strconv.FormatInt(seed, 36) + "~" + strconv.FormatInt(capacity, 36),
		}
	}
	if offset != 0 {
		prev = &models.ListOptions{
			Count:  opts.Count,
			Cursor: strconv.Itoa(offset-opts.Count) + "~" + strconv.FormatInt(seed, 36) + "~" + strconv.FormatInt(capacity, 36),
		}
	}

	// Filter tags to popular / unfiltered tags
	if err := s.filterTags(ctx, data); err != nil {
		return nil, err
	}

	return &models.ListResult{
		Posts: data,
		Next:  next,
		Prev:  prev,
	}, nil
}

func (s *Store) List(rctx context.Context, opts *models.ListOptions) (*models.ListResult, error) {
	ctx, span := trace.StartSpan(rctx, "store.Random")
	defer span.End()

	// Prepare the query based on the opts provided
	q := datastore.NewQuery(models.POST).Limit(opts.Count).Order("-date")
	if opts.Tag != "" {
		q = q.Filter("tags =", opts.Tag)
	}
	if len(opts.Cursor) > 1 {
		cursor, err := datastore.DecodeCursor(opts.Cursor[1:])
		if err != nil {
			return nil, err
		}
		if dir := opts.Cursor[0]; dir == 'e' {
			q = q.End(cursor)
		} else if dir == 's' {
			q = q.Start(cursor)
		} else {
			return nil, errors.New("expected first char in cursor to be 's' or 'e'")
		}
	}
	result := &models.ListResult{}
	iter := s.store.Run(ctx, q)

	// Load the previous cursor (calling back should set q.End to the start of this query)
	if c, err := iter.Cursor(); err != nil {
		return nil, err
	} else if cs := c.String(); cs != `` {
		result.Prev = &models.ListOptions{
			Cursor: "e" + cs,
			Count:  opts.Count,
			Tag:    opts.Tag,
		}
		result.Self = &models.ListOptions{
			Cursor: "s" + cs,
			Count:  opts.Count,
			Tag:    opts.Tag,
		}
	}

	// Load the requested posts
	for {
		var p models.Post
		_, err := iter.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		result.Posts = append(result.Posts, p)
	}

	// Load the next cursor (calling back should set the q.Start to the end of this query)
	if c, err := iter.Cursor(); err != nil {
		return nil, err
	} else if cs := c.String(); cs != `` && len(result.Posts) == opts.Count {
		result.Next = &models.ListOptions{
			Cursor: "s" + cs,
			Count:  opts.Count,
			Tag:    opts.Tag,
		}
		result.Self = &models.ListOptions{
			Cursor: "e" + cs,
			Count:  opts.Count,
			Tag:    opts.Tag,
		}
	}

	// Filter tags to popular / unfiltered tags
	if err := s.filterTags(ctx, result.Posts); err != nil {
		return nil, err
	}

	// Everybody loves metadata
	span.AddAttributes(
		trace.Int64Attribute("posts", int64(len(result.Posts))),
		trace.BoolAttribute("next", result.Next != nil),
		trace.BoolAttribute("prev", result.Prev != nil),
	)
	return result, nil
}

func (s *Store) Has(rctx context.Context, id int64) (bool, error) {
	if s.stash != nil {
		return s.stash[id], nil
	}
	ctx, span := trace.StartSpan(rctx, "store.Has")
	defer span.End()
	keys, err := s.getKeys(ctx, s.store, models.POST)
	if err != nil {
		return false, err
	}
	s.stash = make(map[int64]bool, len(keys))
	for _, key := range keys {
		s.stash[key.ID] = true
	}
	span.AddAttributes(trace.Int64Attribute("keys", int64(len(keys))))
	return s.stash[id], nil
}

func (s *Store) PutMulti(rctx context.Context, posts []models.Post) error {
	if len(posts) == 0 {
		return nil
	}
	ctx, span := trace.StartSpan(rctx, "store.PutMulti")
	defer span.End()
	span.AddAttributes(trace.Int64Attribute("posts", int64(len(posts))))
	keys := make([]*datastore.Key, len(posts))
	for i, post := range posts {
		keys[i] = datastore.IDKey(models.POST, post.ID, nil)
	}
	complete, err := s.store.PutMulti(ctx, keys, posts)
	if err != nil {
		return err
	}
	if s.stash != nil {
		for _, post := range posts {
			s.stash[post.ID] = true
		}
	}
	if err = updateTags(ctx, s.store, posts); err != nil {
		return err
	}
	return addKeys(ctx, s.store, models.POST, complete)
}

func (s *Store) Put(ctx context.Context, post *models.Post) error {
	return s.PutMulti(ctx, []models.Post{*post})
}

func (s *Store) Get(rctx context.Context, id int64) (*models.Post, error) {
	ctx, span := trace.StartSpan(rctx, "store.PutMulti")
	defer span.End()
	span.AddAttributes(trace.Int64Attribute("id", id))

	post := models.Post{} // allocation!
	err := s.store.Get(ctx, datastore.IDKey(models.POST, id, nil), &post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}
