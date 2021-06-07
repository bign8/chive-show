package models

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"regexp"
	"time"

	"cloud.google.com/go/datastore"
)

// POST table to store posts in
const POST = "Post"

// Img the model for a post in an image
type Media struct {
	URL      string `datastore:"url" json:"url" xml:"url,attr"`
	Title    string `datastore:"title" json:"title,omitempty" xml:"-"` // the xml titles are worthless (especially now that the full content isn't present)
	Rating   string `datastore:"-" json:"-" xml:"rating"`
	Category string `datastore:"-" json:"-" xml:"category"`
	// Type     string `datastore:"type" json:"type" xml:"-"` // loaded when scraping content: attachment (for images), gif (for videos)
	Caption string `json:"caption,omitempty"` // .gallery-caption when scraping page content
}

// Post data object
type Post struct {
	ID      int64     `datastore:"-" json:"id,omitempty"`
	Tags    []string  `datastore:"tags" json:"tags"`
	Link    string    `datastore:"link" json:"link"`
	Date    time.Time `datastore:"date" json:"date"`
	Title   string    `datastore:"title" json:"title"`
	Creator string    `datastore:"creator" json:"creator"`
	MugShot string    `datastore:"mugshot,noindex" json:"mugshot"`

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
	return json.Unmarshal(x.MediaBytes, &x.Media)
}

// Save Datastore PropertyLoadSaver Interface : https://pkg.go.dev/cloud.google.com/go/datastore#PropertyLoadSaver
func (x *Post) Save() (props []datastore.Property, err error) {
	if x.MediaBytes, err = json.Marshal(&x.Media); err != nil {
		return nil, err
	}
	return datastore.SaveStruct(x)
}

var clean = regexp.MustCompile(`\W\(([^\)]*)\)$`)

// UnmarshalXML implements xml.Unmarshaler for custom unmarshaling logic : https://golang.org/pkg/encoding/xml/#Unmarshaler
func (x *Post) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var post struct {
		ID      int64    `xml:"post-id"`
		Tags    []string `xml:"category"`
		Link    string   `xml:"link"`
		Date    string   `xml:"pubDate"`
		Title   string   `xml:"title"`
		Creator string   `xml:"creator"`
		Media   []Media  `xml:"content"`
	}
	if err = d.DecodeElement(&post, &start); err != nil {
		return err
	}
	x.ID = post.ID
	x.Tags = post.Tags
	x.Link = post.Link
	x.Title = clean.ReplaceAllLiteralString(post.Title, "")
	x.Creator = post.Creator
	x.Media = post.Media // TODO: do mugshot scrubbing here
	if x.Date, err = time.Parse(time.RFC1123Z, post.Date); err != nil {
		return err
	}
	return nil
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
	Has(ctx context.Context, post Post) (bool, error)
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
