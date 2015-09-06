package main

import (
  "encoding/json"
  "fmt"
  "math/rand"
  "net/http"
  "net/url"
  "strconv"

  "helpers/keycache"

  "appengine"
  "appengine/datastore"

  "github.com/mjibson/appstats"
)

func api() {
  http.Handle("/api/v1/post/random", appstats.NewHandler(random))
}

// API Helper function
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

// Actual PI functions
func random(c appengine.Context, w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  count := get_url_count(r.URL)
  c.Infof("Requested %v random posts", count)

  result := JsonPostResponse{
    Status: "error",
    Code:   500,
  }

  // Pull keys from post keys object
  keys, err := keycache.GetKeys(c, DB_POST_TABLE)
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
