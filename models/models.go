package models

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"time"

	"cloud.google.com/go/datastore"
)

// POST table to store posts in
const POST = "Post"

// Img the model for a post in an image
type Media struct {
	ID      int64  `json:"id,omitempty"`      // used for matching media metadata to attachment ordering
	URL     string `json:"url"`               // literal link for the asset in question
	Title   string `json:"title,omitempty"`   // the xml titles are worthless (especially now that the full content isn't present)
	Caption string `json:"caption,omitempty"` // .gallery-caption when scraping page content
	Height  int    `json:"height,omitempty"`
	Width   int    `json:"width,omitempty"`
}

// Post data object
type Post struct {
	ID      int64     `datastore:"-" json:"id,omitempty"`
	Tags    []string  `datastore:"tags" json:"tags"`
	Link    string    `datastore:"link,noindex" json:"link"`
	Date    time.Time `datastore:"date" json:"date"`
	Title   string    `datastore:"title,noindex" json:"title"`
	Creator string    `datastore:"creator,noindex" json:"creator"`
	MugShot string    `datastore:"mugshot,noindex" json:"mugshot"`
	Thumb   string    `datastore:"thumbnail,noindex" json:"thumbnail"`

	// What version of the miner was used to scrape this post together?
	Version int `datastore:"version" json:"version"`

	// Attributes tweaked to minimize transactions (LoadSaver stuff)
	MediaBytes []byte  `datastore:"media,noindex" json:"-"`
	Media      []Media `datastore:"-" json:"media"`
}

// LoadKey Datastore KeyLoader Interface: https://pkg.go.dev/cloud.google.com/go/datastore#KeyLoader
func (x *Post) LoadKey(k *datastore.Key) error {
	x.ID = k.ID
	return nil
}

// Load Datastore PropertyLoadSaver Interface : https://pkg.go.dev/cloud.google.com/go/datastore#PropertyLoadSaver
func (x *Post) Load(c []datastore.Property) error {
	err := datastore.LoadStruct(x, c)
	if err != nil {
		return err
	}
	cmp, err := gzip.NewReader(bytes.NewReader(x.MediaBytes))
	if err != nil {
		return err
	}
	if err = json.NewDecoder(cmp).Decode(&x.Media); err != nil {
		return err
	}
	x.MediaBytes = nil
	return cmp.Close()
}

// Save Datastore PropertyLoadSaver Interface : https://pkg.go.dev/cloud.google.com/go/datastore#PropertyLoadSaver
func (x *Post) Save() ([]datastore.Property, error) {
	var buffer bytes.Buffer
	stream := gzip.NewWriter(&buffer)
	enc := json.NewEncoder(stream)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(&x.Media); err != nil {
		return nil, err
	}
	if err := stream.Close(); err != nil {
		return nil, err
	}
	x.MediaBytes = buffer.Bytes()
	return datastore.SaveStruct(x)
}

// Storage based errors
// TODO: wrap the base errors better
var (
	ErrNotEnough = errors.New("not enough")
)

// Store defines a general abstraction over storage operations
type Store interface {
	Random(ctx context.Context, opts *ListOptions) (*ListResult, error)
	List(ctx context.Context, opts *ListOptions) (*ListResult, error)
	Tags(ctx context.Context) (map[string]int, error) // name => len(posts)
	// Delete(ctx context.Context, ids []int) error
	PutMulti(ctx context.Context, posts []Post) error
	Has(ctx context.Context, id int64) (bool, error)
	Get(ctx context.Context, id int64) (*Post, error)
	Put(ctx context.Context, post *Post) error
}

type ListOptions struct {
	Count  int
	Cursor string
	Tag    string
}

type ListResult struct {
	Posts []Post
	Next  *ListOptions
	Prev  *ListOptions
	Self  *ListOptions
}
