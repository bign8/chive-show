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
	keyz, err := db.GetAll(context.Background(), datastore.NewQuery(`Post`), &posts)
	if err != nil {
		panic(err)
	}

	// Rewrite all the posts in case there was a schema update
	_, err = db.PutMulti(context.Background(), keyz, posts)
	if err != nil {
		panic(err)
	}

	// Update all the TAGS
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
	keys := make([]*datastore.Key, 0, len(tags))
	for tag := range tags {
		keys = append(keys, tagKey(tag))
	}
	_, err := client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		tagz := make([]Tag, len(keys))
		err := tx.GetMulti(keys, tagz)
		err = ignoreNoneSuchEntity(err)
		if err != nil {
			return err
		}
		for i, key := range keys {
			tagz[i].Count += tags[key.Name]
		}
		_, err = tx.PutMulti(keys, tagz)
		return err
	})
	return err
}

func ignoreNoneSuchEntity(err error) error {
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	if merr, ok := err.(datastore.MultiError); ok {
		var real bool
		for _, e := range merr {
			if e != nil && e != datastore.ErrNoSuchEntity {
				return err
			}
		}
		if !real {
			return nil
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
