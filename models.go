package api

import "appengine/datastore"

type Author struct {
  Name string `datastore:"name" json:"name"`
  Img  string `datastore:"img" json:"img"`
}

type Tag struct {
  Name string `datastore:"name" json:"name"`
}

type Img struct {
  Url      string `datastore:"url" json:"url"`
  Title    string `datastore:"title" json:"title"`
  IsValid  bool   `datastore:"is_valid" json:"is_valid"`
  Rating   string `datastore:"rating"`
}

type Post struct {
  Tags    []string         `datastore:"tags" json:"tags"`
  Link    string           `datastore:"link" json:"link"`
  Date    string           `datastore:"date" json:"date"`
  Title   string           `datastore:"title" json:"title"`
  Author  *datastore.Key   `datastore:"author" json:"creator"`
  Imgs    []*datastore.Key `datastore:"keys" json:"media"`
  Media   []byte           `datastore:"media"`
  Creator []byte           `datastore:"creator"`
  Guid    string           `datastore:"guid"`
}

type PostKeys struct {
  Keys  []PostKey
}

type PostKey struct {
  StringID  string
  IntID     int64
}

type JsonPostResponse struct {
  Status string     `json:"status"`
  Code   int        `json:"code"`
  Data   []JsonPost `json:"data"`
}

type JsonPost struct {
  Tags    []string `json:"tags"`
  Link    string   `json:"link"`
  Date    string   `json:"date"`
  Title   string   `json:"title"`
  Author  Author   `json:"creator"`
  Imgs   []Img     `json:"media"`
}
