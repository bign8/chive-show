package models

type Img struct {
  Url      string `datastore:"url"      json:"url"   xml:"url,attr"`
  Title    string `datastore:"title"    json:"title" xml:"title"`
  Rating   string `datastore:"rating"   json:"-"     xml:"rating"`
  Category string `datastore:"-"        json:"-"     xml:"category"`

  // TODO: Implement these!
  IsValid  bool   `datastore:"is_valid" json:"-"     xml:"-"`
}
