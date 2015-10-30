package models

// Img the model for a post in an image
type Img struct {
	URL      string `datastore:"url"      json:"url"   xml:"url,attr"`
	Title    string `datastore:"title"    json:"title" xml:"title"`
	Rating   string `datastore:"rating"   json:"-"     xml:"rating"`
	Category string `datastore:"-"        json:"-"     xml:"category"`

	// TODO: Implement these!
	IsValid bool `datastore:"is_valid" json:"-"     xml:"-"`
}
