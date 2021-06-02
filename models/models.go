package models

import (
	"context"
	"encoding/json"
	"errors"

	"cloud.google.com/go/datastore"
)

// POST table to store posts in
const POST = "Post"

// Img the model for a post in an image
type Img struct {
	URL      string `datastore:"url"      json:"url"   xml:"url,attr"`
	Title    string `datastore:"title"    json:"title" xml:"title"`
	Rating   string `datastore:"rating"   json:"-"     xml:"rating"`
	Category string `datastore:"-"        json:"-"     xml:"category"`

	// TODO: Implement these!
	IsValid bool `datastore:"is_valid" json:"-"     xml:"-"`
}

// Post data object
type Post struct {
	// Attributes used for generating unique leu
	GUID string `datastore:"-"       json:"-"       xml:"guid"`

	// Direct atributes
	Tags    []string `datastore:"tags"    json:"tags"    xml:"category"`
	Link    string   `datastore:"link"    json:"link"    xml:"link"`
	Date    string   `datastore:"date"    json:"date"    xml:"pubDate"`
	Title   string   `datastore:"title"   json:"title"   xml:"title"`
	Creator string   `datastore:"creator" json:"creator" xml:"creator"`

	// Attributes tweaked to minimize transactions (LoadSaver stuff)
	MediaBytes []byte `datastore:"media,noindex"   json:"-"       xml:"-"`
	Media      []Img  `datastore:"-"               json:"media"   xml:"content"`

	// Manually built attributes from post
	MugShot string `datastore:"mugshot" json:"mugshot" xml:"-"`
}

// Load Datastore LoadSaveProperty Interface
func (x *Post) Load(c []datastore.Property) error {
	if err := datastore.LoadStruct(x, c); err != nil {
		return err
	}
	return json.Unmarshal(x.MediaBytes, &x.Media)
}

// Save Datastore LoadSaveProperty Interface
func (x *Post) Save() (props []datastore.Property, err error) {
	if x.MediaBytes, err = json.Marshal(&x.Media); err != nil {
		return nil, err
	}
	return datastore.SaveStruct(x)
}

// Storage based errors
// TODO: wrap the base errors better
var (
	ErrNotEnough = errors.New("not enough")
)

// Store defines a general abstraction over storage operations
type Store interface {
	Random(ctx context.Context, opts *RandomOptions) (*RandomResult, error)
	// List(ctx context.Context, opts *ListOptions) ([]Post, error)
	// Save(ctx context.Context, post Post) error
	// Create(ctx context.Context, posts []Post) error
	// Delete(ctx context.Context, ids []int) error
}

type RandomOptions struct {
	Count  int
	Cursor string
}

type RandomResult struct {
	Posts []Post
	Next  *RandomOptions
	Prev  *RandomOptions
}
