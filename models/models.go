package models

import (
	"encoding/json"

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
	MediaBytes []byte `datastore:"media"   json:"-"       xml:"-"`
	Media      []Img  `datastore:"-"       json:"media"   xml:"content"`

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
