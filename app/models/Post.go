package models

import (
  "appengine/datastore"
  "encoding/json"
)

type Post struct {
  // Attributes used for generating unique leu
  Guid       string   `datastore:"-"       json:"-"       xml:"guid"`

  // Direct atributes
  Tags       []string `datastore:"tags"    json:"tags"    xml:"category"`
  Link       string   `datastore:"link"    json:"link"    xml:"link"`
  Date       string   `datastore:"date"    json:"date"    xml:"pubDate"`
  Title      string   `datastore:"title"   json:"title"   xml:"title"`
  Creator    string   `datastore:"creator" json:"creator" xml:"creator"`

  // Attributes tweaked to minimize transactions (LoadSaver stuff)
  MediaBytes []byte   `datastore:"media"   json:"-"       xml:"-"`
  Media      []Img    `datastore:"-"       json:"media"   xml:"content"`

  // Manually built attributes from post
  MugShot    string   `datastore:"mugshot" json:"mugshot" xml:"-"`
}

// Datastore LoadSaveProperty Interface
func (x *Post) Load(c <-chan datastore.Property) error {
  if err := datastore.LoadStruct(x, c); err != nil {
    return err
  }
  return json.Unmarshal(x.MediaBytes, &x.Media)
}

func (x *Post) Save(c chan<- datastore.Property) (err error) {
  if x.MediaBytes, err = json.Marshal(&x.Media); err != nil {
    close(c)
    return err
  }
  return datastore.SaveStruct(x, c)
}
