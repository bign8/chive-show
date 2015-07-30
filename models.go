package api

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
}

type Post struct {
  Tags    []string `datastore:"tags" json:"tags"`
  Link    string   `datastore:"link" json:"link"`
  Date    string   `datastore:"date" json:"date"`
  Title   string   `datastore:"title" json:"title"`
  Author  Author   `datastore:"author" json:"creator"`
  Imgs   []Img     `datastore:"keys" json:"media"`
}

type PostResponse struct {
  Status string `json:"status"`
  Code   int    `json:"code"`
  Data   []Post `json:"data"`
}
