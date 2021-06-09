package datastore

import (
	"context"
	"log"
	"strings"

	"cloud.google.com/go/datastore"
	"go.opencensus.io/trace"

	"github.com/bign8/chive-show/models"
)

type Tag struct {
	Count int `datastore:"count"`
}

// transactions involving multiple keys need to have the same parent
var tagParent = datastore.NameKey(`Site`, `thechive`, nil)

func tagKey(name string) *datastore.Key {
	return datastore.NameKey(`Tag`, name, tagParent)
}

func (s *Store) Tags(rctx context.Context) (map[string]int, error) {
	ctx, span := trace.StartSpan(rctx, "store.Tags")
	defer span.End()
	var tags []Tag
	keys, err := s.store.GetAll(ctx, datastore.NewQuery(`Tag`).Filter("count >", 3), &tags)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil // prevent divide by 0 below
	}

	// Dynamic culling of list (removing the "outliers" of <= 3; not mathematically sound, but :shrug:)
	var total int
	for _, tag := range tags {
		total += tag.Count
	}
	total /= len(tags) // compute the average
	total /= 2         // very rough 75th percentile

	out := make(map[string]int, len(tags))
	for i, tag := range tags {
		if tag.Count < total {
			continue // don't bother with smalls
		}
		name := keys[i].Name
		if strings.HasSuffix(name, "staple") || name == "full" {
			continue // cull after the fact so we can change logic and not have to migrate data
		}
		out[name] = tag.Count
	}
	span.AddAttributes(trace.Int64Attribute("tags", int64(len(out))))
	return out, nil
}

func (s *Store) filterTags(ctx context.Context, posts []models.Post) error {
	// TODO: memory cache of s.Tags response
	tags, err := s.Tags(ctx)
	if err != nil {
		return err
	}
	for i, post := range posts {
		ts := post.Tags
		for j := len(ts) - 1; j >= 0; j-- {
			if _, ok := tags[ts[j]]; !ok {
				post.Tags = append(post.Tags[:j], post.Tags[j+1:]...)
			}
		}
		posts[i] = post
	}
	return nil
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
