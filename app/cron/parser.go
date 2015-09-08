package cron

import (
  // "app/models"
  // "app/helpers/keycache"
  // "appengine"
  // "appengine/datastore"
  // "appengine/delay"
  // "appengine/taskqueue"
  // "appengine/urlfetch"
  "encoding/xml"
  // "encoding/json"
  // "fmt"
  // "net/http"
  "html/template"
)

type Node struct {
  // XML string `xml:",innerxml"`
  // ATTR []string
  // DATA string `xml:",chardata"`
  XMLName xml.Name
  XMLAttrs []xml.Attr `xml:",any"`
  DATA string `xml:",chardata"`
}

type Post struct {
  Guid       string   `xml:"guid"`
  Tags       []string `xml:"category"`
  Link       string   `xml:"link"`
  Date       string   `xml:"pubDate"`
  Title      string   `xml:"title"`
  Creator    string   `xml:"creator"`
  Media      []Img    `xml:"content"`
  CommentRSS string   `xml:"commentRss"`
  Comment    []string   `xml:"comments"`
  Desc       template.HTML   `xml:"description"`
  Enclosure  struct {
    Url      string   `xml:"url,attr"`
    Children []Node   `xml:",any"`
  }   `xml:"enclosure"`
  Thumbnail  struct {
    Url      string   `xml:"url,attr"`
    Children []Node   `xml:",any"`
  }   `xml:"thumbnail"`
  Children   []Node   `xml:",any"`
  Content    template.HTML  `xml:"encoded"`
}

type Img struct {
  Url      string `xml:"url,attr"`
  Title    string `xml:"title"`
  Rating   string `xml:"rating"`
  Category string `xml:"category"`
}

// Worker: this will be a worker on defered work chains

func parseData(data string) (*Post, error) {
  var post Post
  err := xml.Unmarshal([]byte(data), &post)
  return &post, err
}
