package datastore

import (
	"context"
	"log"

	"cloud.google.com/go/datastore"
	"go.opencensus.io/trace"

	"github.com/bign8/chive-show/models"
)

type Tag struct {
	Count int `datastore:"count"`
}

func tagKey(name string) *datastore.Key { return datastore.NameKey(`Tag`, name, nil) }

func (s *Store) Tags(rctx context.Context) (map[string]int, error) {
	ctx, span := trace.StartSpan(rctx, "store.Tags")
	defer span.End()
	var tags []Tag
	keys, err := s.store.GetAll(ctx, datastore.NewQuery(`Tag`).Filter("count >", 3), &tags) // TODO: dynamically set the limit based on the DB size and expected frequency
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(tags))
	for i, tag := range tags {
		out[keys[i].Name] = tag.Count
	}
	span.AddAttributes(trace.Int64Attribute("tags", int64(len(out))))
	return out, nil
}

func rebuildTags(db *datastore.Client) {
	var posts []models.Post
	_, err := db.GetAll(context.Background(), datastore.NewQuery(`Post`), &posts)
	if err != nil {
		panic(err)
	}
	tags := posts2tagMap(posts)
	keys := make([]*datastore.Key, 0, len(tags))
	list := make([]Tag, 0, len(tags))
	for name, count := range tags {
		keys = append(keys, tagKey(name))
		list = append(list, Tag{Count: count})
	}
	_, err = db.PutMulti(context.Background(), keys, list)
	if err != nil {
		panic(err)
	}
	log.Printf("INFO: Rebuilt %d tags", len(tags))
}

func updateTags(ctx context.Context, client datastoreClient, posts []models.Post) error {
	tags := posts2tagMap(posts)
	var errz datastore.MultiError
	for name, count := range tags {
		if err := incrementTag(ctx, name, count, client); err != nil {
			errz = append(errz, err)
			log.Printf("WARN: Failed to increment tag %q by %d", name, count)
		}
	}
	if len(errz) != 0 {
		return errz
	}
	return nil
}

// counter example from: https://pkg.go.dev/cloud.google.com/go/datastore#Client.NewTransaction
func incrementTag(ctx context.Context, name string, count int, client datastoreClient) (err error) {
	key := tagKey(name)
	var tx *datastore.Transaction
	for i := 0; i < 3; i++ {
		tx, err = client.NewTransaction(ctx)
		if err != nil {
			break
		}

		var c Tag
		if err = tx.Get(key, &c); err != nil && err != datastore.ErrNoSuchEntity {
			break
		}
		c.Count += count
		if _, err = tx.Put(key, &c); err != nil {
			break
		}

		// Attempt to commit the transaction. If there's a conflict, try again.
		if _, err = tx.Commit(); err != datastore.ErrConcurrentTransaction {
			break
		}
	}
	return err
}

func posts2tagMap(posts []models.Post) map[string]int {
	tags := make(map[string]int)
	for _, post := range posts {
		for _, tag := range post.Tags {
			tags[tag] += 1
		}
	}
	return tags
}
