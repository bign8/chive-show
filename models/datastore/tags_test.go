package datastore

import (
	"context"
	"strconv"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/bign8/chive-show/models"
)

type ignoreNoneSuchEntityCase struct {
	in    error
	clear bool
}

func (tc ignoreNoneSuchEntityCase) run(t *testing.T) {
	err := ignoreNoneSuchEntity(tc.in)
	if tc.clear {
		if err != nil {
			t.Errorf("arg %v; got %v; want nil", tc.in, err)
		}
	} else if err == nil {
		t.Errorf("arg %v; got nil; want inp", tc.in)
	}
}

func TestIgnoreNoneSuchEntity(t *testing.T) {
	for i, test := range []ignoreNoneSuchEntityCase{
		{nil, true},
		{datastore.ErrInvalidKey, false},
		{datastore.ErrNoSuchEntity, true},
		{datastore.MultiError{nil}, true},
		{datastore.MultiError{datastore.ErrNoSuchEntity}, true},
		{datastore.MultiError{datastore.ErrInvalidKey}, false},
	} {
		t.Run(strconv.Itoa(i), test.run)
	}
}

func TestUpdateTags(t *testing.T) {
	posts := []models.Post{
		{Tags: []string{"a", "b", "c"}},
		{Tags: []string{"a", "d", "e"}},
		{Tags: []string{"a", "f", "g"}},
	}
	err := updateTags(context.TODO(), &fake{}, posts)
	if err != nil {
		t.Fatalf("UpdateTags Failed: %v", err)
	}
}

func TestTags(t *testing.T) {
	s := &Store{
		store: &fake{
			getAll: func(obj interface{}) ([]*datastore.Key, error) {
				*(obj.(*[]Tag)) = []Tag{{Count: 3}}
				return []*datastore.Key{
					datastore.NameKey(``, `asdf`, nil),
				}, nil
			},
		},
	}
	tags, err := s.Tags(context.TODO())
	if err != nil {
		t.Fatalf("store.Tags Failed: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
	for tag, count := range tags {
		if tag != "asdf" {
			t.Errorf("Expected tag asdf; got %s", tag)
		}
		if count != 3 {
			t.Errorf("Expected count 3; got %d", count)
		}
	}
}
