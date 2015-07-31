package api

import (
  "encoding/json"
  "fmt"
  "net/http"

  "appengine"
  "appengine/datastore"
)

func init() {
  http.HandleFunc("/", http.NotFound)  // Default Handler too
  http.HandleFunc("/api/v1/post/random", random)
  http.HandleFunc("/api/load", load)
}

func random(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  w.Header().Set("Content-Type", "application/json")

  // Pull posts from datastore
  var data []Post
  q := datastore.NewQuery("Post").Limit(5)
  if _, err := q.GetAll(c, &data); err != nil {
      c.Errorf("datastore.NewQuery %v", err)
      fmt.Fprint(w, "{\"status\":\"error\",\"code\":500,\"data\":null}")
      return
  }

  // Convert posts to json and pull linked assets
  json_data := make([]JsonPost, len(data))
  for idx, item := range data {
    imgs := make([]Img, len(item.Imgs))
    err := datastore.GetMulti(c, item.Imgs, imgs)
    if err != nil {
      c.Errorf("datastore.GetMulti %v", err)
      imgs = nil
    }
    var author Author
    if err := json.Unmarshal(item.Creator, &author); err != nil {
      c.Errorf("json.Unmarshal %v", err)
      author = Author{Name:"Unknown", Img:"http://www.clker.com/cliparts/5/9/4/c/12198090531909861341man%20silhouette.svg.hi.png"}
    }
    json_data[idx] = JsonPost{
      Tags: item.Tags,
      Link: item.Link,
      Date: item.Date,
      Title: item.Title,
      Author: author,
      Imgs: imgs,
    }
  }
  result := &JsonPostResponse{Status: "success", Code: 200, Data: json_data}

  str_items, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
    c.Errorf("json.MarshalIndent %v", err)
    fmt.Fprint(w, "{\"status\":\"error\",\"code\":500,\"data\":null}")
    return
  }
  fmt.Fprint(w, string(str_items))
}

func load(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  inno_key := datastore.NewIncompleteKey(c, "Post", nil)

  img_key_1, err := datastore.Put(c,
    datastore.NewIncompleteKey(c, "Img", nil),
    &Img{
      Url: "https://thechive.files.wordpress.com/2015/02/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-10.jpg",
      Title: "\u201cI swear doc, I don\u2019t know how it got there\u201d",
      IsValid: true,
    },
  )
  if err != nil {
    fmt.Print(w, "Error with img_key_1")
  }

  obj := &Post{
    Tags: []string{"Awesome", "Funny", "of", "posts", "the", "top", "week"},
    Link: "http://thechive.com/2015/02/15/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-2/",
    Date: "Sun, 15 Feb 2015 22:48:43 +0000",
    Title: "In case you missed them, check out the Top Posts of the Week (10 Photos)",
    Creator: []byte("{\"name\": \"Dougy\", \"img\": \"http://1.gravatar.com/avatar/1e03788f973939e10eb6cf27e644c78a?s=50\u0026d=http%3A%2F%2F1.gravatar.com%2Favatar%2Fad516503a11cd5ca435acc9bb6523536%3Fs%3D50\u0026r=X\"}"),
    Author: nil,
    Imgs: []*datastore.Key{
      img_key_1,
    },
  }
  _, err = datastore.Put(c, inno_key, obj)
  if err != nil {
    c.Errorf("datastore.Put %v", err)
  }
}
