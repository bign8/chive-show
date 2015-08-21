package api

import (
  "encoding/json"
  "fmt"
  "math/rand"
  "net/http"
  "net/url"
  "strconv"

  "helpers"

  "appengine"
  "appengine/datastore"

  "github.com/mjibson/appstats"
)

func init() {
  http.HandleFunc("/", http.NotFound)  // Default Handler too
  http.Handle("/api/v1/post/random", appstats.NewHandler(random))
  http.HandleFunc("/api/load", load)
  http.HandleFunc("/cron/parse_feeds", parseFeeds)
}

func parseFeeds(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, "Here is your cron")
}

func get_url_count(url *url.URL) int {
  x := url.Query().Get("count")
  if x == "" {
    return 2
  }
  val, err := strconv.Atoi(x)
  if err != nil || val > 30 || val < 1 {
    return 2
  }
  return val
}

func random(c appengine.Context, w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  count := get_url_count(r.URL)
  c.Infof("Requested %v random posts", count)

  result := JsonPostResponse{
    Status: "error",
    Code:   500,
  }

  // Pull keys from post keys object
  keys, err := helpers.GetKeys(c, "Post")
  if err != nil {

    c.Errorf("heleprs.GetKeys %v", err)
    result.Msg = "Error with helpers GetKeys"

  } else if len(keys) < count {

    c.Errorf("Not enough keys(%v) for count(%v)", len(keys), count)
    result.Msg = "Basically empty datastore"

  } else {

    // Randomize list of keys
    for i := range keys {
      j := rand.Intn(i + 1)
      keys[i], keys[j] = keys[j], keys[i]
    }

    // Pull posts from datastore
    data := make([]Post, count)
    if err := datastore.GetMulti(c, keys[:count], data); err != nil {

      c.Errorf("datastore.GetMulti %v", err)
      result.Msg = "Error with datastore GetMulti"

    } else {

      // Generate JsonPostResponse object
      result.Status = "success"
      result.Code = 200
      result.Msg = "Your amazing data awaits"
      result.Data = data
    }
  }

  // Serialize and send response
  str_items, err := json.MarshalIndent(&result, "", "  ")
  var out string
  if err != nil {
    c.Errorf("json.MarshalIndent %v", err)
    out = "{\"status\":\"error\",\"code\":500,\"data\":null,\"msg\":\"Error marshaling data\"}"
  } else {
    out = string(str_items)
  }
  fmt.Fprint(w, out)
}

func load(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  inno_key := datastore.NewIncompleteKey(c, "Post", nil)

  image := Img{
    Url: "https://thechive.files.wordpress.com/2015/02/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-10.jpg",
    Title: "\u201cI swear doc, I don\u2019t know how it got there\u201d",
    IsValid: true,
  }

  img_key_1, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Img", nil), &image)
  if err != nil {
    fmt.Print(w, "Error with img_key_1")
  }

  b, err := json.Marshal([]Img{image, image})
  if err != nil {
    fmt.Print(w, "Error with Marshaling")
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
    Media: b,
  }
  _, err = datastore.Put(c, inno_key, obj)
  if err != nil {
    c.Errorf("datastore.Put %v", err)
  }
}
