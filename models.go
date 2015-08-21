package api

import (
  "appengine/datastore"
  "encoding/json"
)

type Author struct {
  Name string `datastore:"name" json:"name"`
  Img  string `datastore:"img"  json:"img"`
}

type Img struct {
  Url      string `datastore:"url"      json:"url"`
  Title    string `datastore:"title"    json:"title"`
  IsValid  bool   `datastore:"is_valid" json:"-"`
  Rating   string `datastore:"rating"   json:"-"`
}

type Post struct {
  Tags      []string         `datastore:"tags"    json:"tags"`
  Link      string           `datastore:"link"    json:"link"`
  Date      string           `datastore:"date"    json:"date"`
  Title     string           `datastore:"title"   json:"title"`

  Author    *datastore.Key   `datastore:"author"  json:"-"`
  Imgs      []*datastore.Key `datastore:"keys"    json:"-"`
  Media     []byte           `datastore:"media"   json:"-"`
  Creator   []byte           `datastore:"creator" json:"-"`
  Guid      string           `datastore:"guid"    json:"-"`

  JsCreator Author           `datastore:"-"       json:"creator"`
  JsImgs    []Img            `datastore:"-"       json:"media"`
}

func (x *Post) Load(c <-chan datastore.Property) error {
    if err := datastore.LoadStruct(x, c); err != nil {
        return err
    }
    // Load Author
    if json.Unmarshal(x.Creator, &x.JsCreator) != nil {
      x.JsCreator = Author{
        Name: "Unknown",
        Img: "http://www.clker.com/cliparts/5/9/4/c/12198090531909861341man%20silhouette.svg.hi.png",
      }
    }
    // Load Images/Media
    return json.Unmarshal(x.Media, &x.JsImgs)
}

func (x *Post) Save(c chan<- datastore.Property) error {
    // defer close(c)
    return datastore.SaveStruct(x, c)
}

type JsonPostResponse struct {
  Status string     `json:"status"`
  Code   int        `json:"code"`
  Data   []Post     `json:"data"`
  Msg    string     `json:"msg"`
}
