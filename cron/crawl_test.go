package cron

import (
	"context"
	"testing"

	"github.com/bign8/chive-show/models"
)

type notHasStore struct {
	models.Store
}

func (notHasStore) Has(context.Context, int64) (bool, error) { return false, nil }

func TestGetAndParseFeed(t *testing.T) {
	_, posts, err := getAndParseFeed(context.Background(), notHasStore{}, 1)
	if err != nil {
		t.Fatalf("Unable to get and parse feed: %v", err)
	}
	t.Logf("TODO: make assertions on %v", posts)
}
