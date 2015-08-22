package main

import "appengine/datastore"

type Author struct {
  Name string `datastore:"name" json:"name"`
  Img  string `datastore:"img"  json:"img"`
}

type Img struct {
  Url      string `datastore:"url"      json:"url"    xml:"url,attr"`
  Title    string `datastore:"title"    json:"title"  xml:"title"`
  IsValid  bool   `datastore:"is_valid" json:"-"      xml:"-"`
  Rating   string `datastore:"rating"   json:"-"      xml:"rating"`
  Category string `datastore:"-"        json:"-"      xml:"category"`
}

type Post struct {
  Tags      []string         `datastore:"tags"    json:"tags"    xml:"category"`
  Link      string           `datastore:"link"    json:"link"    xml:"link"`
  Date      string           `datastore:"date"    json:"date"    xml:"pubDate"`
  Title     string           `datastore:"title"   json:"title"   xml:"title"`

  Author    *datastore.Key   `datastore:"author"  json:"-"       xml:"-"`
  Imgs      []*datastore.Key `datastore:"keys"    json:"-"       xml:"-"`
  Media     []byte           `datastore:"media"   json:"-"       xml:"-"`
  Creator   []byte           `datastore:"creator" json:"-"       xml:"-"`
  Guid      string           `datastore:"guid"    json:"-"       xml:"guid"`

  JsCreator Author           `datastore:"-"       json:"creator" xml:"-"`
  JsImgs    []Img            `datastore:"-"       json:"media"   xml:"content"`

  StrAuthor string           `datastore:"-"       json:"-"       xml:"creator"`
}

type JsonPostResponse struct {
  Status string     `json:"status"`
  Code   int        `json:"code"`
  Data   []Post     `json:"data"`
  Msg    string     `json:"msg"`
}
