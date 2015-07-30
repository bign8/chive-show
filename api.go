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
  http.HandleFunc("/api/get", read)
}

func read(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  w.Header().Set("Content-Type", "application/json")

  result := &PostResponse{Status: "success", Code: 200, Data: nil}
  datastore.NewQuery("Post").GetAll(c, &result.Data)

  str_items, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
    c.Errorf("json.MarshalIndent %v", err)
    fmt.Fprint(w, "{\"status\":\"error\",\"code\":200,\"data\":null}")
    return
  }
  fmt.Fprint(w, string(str_items))
}

func load(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  inno_key := datastore.NewIncompleteKey(c, "Post", nil)
  obj := &Post{
    Tags: []string{"Awesome", "Funny", "of", "posts", "the", "top", "week"},
    Link: "http://thechive.com/2015/02/15/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-2/",
    Date: "Sun, 15 Feb 2015 22:48:43 +0000",
    Title: "In case you missed them, check out the Top Posts of the Week (10 Photos)",
    Author: Author{
      Name: "Ben",
      Img: "http://0.gravatar.com/avatar/67403a19b8ff2589cad1002324aaad88?s=50&d=http%3A%2F%2F0.gravatar.com%2Favatar%2Fad516503a11cd5ca435acc9bb6523536%3Fs%3D50&r=X",
    },
    Imgs: []Img{
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
      Img{
        Url: "Some URL 2",
        Title: "Some title 2",
        IsValid: true,
      },
      Img{
        Url: "Some URL 2",
        Title: "Some title 2",
        IsValid: true,
      },
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
      Img{
        Url: "Some URL 1",
        Title: "Some title 1",
        IsValid: true,
      },
    },
  }
  _, err := datastore.Put(c, inno_key, obj)
  if err != nil {
    c.Errorf("datastore.Put %v", err)
  }
}

func random(w http.ResponseWriter, r *http.Request) {
	// c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{
    "status": "success", "code": 200, "data": [
    {
    "author": null,
    "title": "In case you missed them, check out the Top Posts of the Week (10 Photos)",
    "media": [
    {"category": [], "rating": null, "title": "\u201cI swear doc, I don\u2019t know how it got there\u201d", "url": "https://thechive.files.wordpress.com/2015/02/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-10.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2coIBCxIDSW1nInlodHRwczovL3RoZWNoaXZlLmZpbGVzLndvcmRwcmVzcy5jb20vMjAxNS8wMi9pbi1jYXNlLXlvdS1taXNzZWQtdGhlbS1jaGVjay1vdXQtdGhlLXRvcC1wb3N0cy1vZi10aGUtd2Vlay0xMC1waG90b3MtMTAuanBnDA", "is_valid": true},
    {"category": [], "rating": null, "title": "Actors that have dipped their toes into the adult film industry", "url": "https://thechive.files.wordpress.com/2015/02/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-7.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2coEBCxIDSW1nInhodHRwczovL3RoZWNoaXZlLmZpbGVzLndvcmRwcmVzcy5jb20vMjAxNS8wMi9pbi1jYXNlLXlvdS1taXNzZWQtdGhlbS1jaGVjay1vdXQtdGhlLXRvcC1wb3N0cy1vZi10aGUtd2Vlay0xMC1waG90b3MtNy5qcGcM", "is_valid": true},
    {"category": [], "rating": null, "title": "If you\u2019re going to cheat, don\u2019t get caught on Facebook", "url": "https://thechive.files.wordpress.com/2015/02/bcb0519c2a5df1b02b3752d307e0dee4_650x11.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cmELEgNJbWciWGh0dHBzOi8vdGhlY2hpdmUuZmlsZXMud29yZHByZXNzLmNvbS8yMDE1LzAyL2JjYjA1MTljMmE1ZGYxYjAyYjM3NTJkMzA3ZTBkZWU0XzY1MHgxMS5qcGcM", "is_valid": true},
    {"category": [], "rating": null, "title": "The gym isn't for everyone...", "url": "https://thechive.files.wordpress.com/2015/02/the-gym-isnt-for-everyone-32-photos-27-e14235868221351.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cnALEgNJbWciZ2h0dHBzOi8vdGhlY2hpdmUuZmlsZXMud29yZHByZXNzLmNvbS8yMDE1LzAyL3RoZS1neW0taXNudC1mb3ItZXZlcnlvbmUtMzItcGhvdG9zLTI3LWUxNDIzNTg2ODIyMTM1MS5qcGcM", "is_valid": true},
    {"category": [], "rating": null, "title": "A few school lunches from around the world", "url": "https://thechive.files.wordpress.com/2015/02/a-few-school-lunches-from-around-the-world-9-photos-12.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cnALEgNJbWciZ2h0dHBzOi8vdGhlY2hpdmUuZmlsZXMud29yZHByZXNzLmNvbS8yMDE1LzAyL2EtZmV3LXNjaG9vbC1sdW5jaGVzLWZyb20tYXJvdW5kLXRoZS13b3JsZC05LXBob3Rvcy0xMi5qcGcM", "is_valid": true},
    {"category": [], "rating": null, "title": "The difference between men and women illustrated perfectly in one written assignment", "url": "https://thechive.files.wordpress.com/2015/02/difference-between-men-women-assignment-0.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cmMLEgNJbWciWmh0dHBzOi8vdGhlY2hpdmUuZmlsZXMud29yZHByZXNzLmNvbS8yMDE1LzAyL2RpZmZlcmVuY2UtYmV0d2Vlbi1tZW4td29tZW4tYXNzaWdubWVudC0wLmpwZww", "is_valid": true},
    {"category": [], "rating": null, "title": "Facebook: The good, the bad, and the ugly", "url": "https://thechive.files.wordpress.com/2015/02/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-5.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2coEBCxIDSW1nInhodHRwczovL3RoZWNoaXZlLmZpbGVzLndvcmRwcmVzcy5jb20vMjAxNS8wMi9pbi1jYXNlLXlvdS1taXNzZWQtdGhlbS1jaGVjay1vdXQtdGhlLXRvcC1wb3N0cy1vZi10aGUtd2Vlay0xMC1waG90b3MtNS5qcGcM", "is_valid": true},
    {"category": [], "rating": null, "title": "If your ex texts you, stay strong and don\u2019t give in", "url": "https://thechive.files.wordpress.com/2015/02/if-your-ex-texts-you-stay-strong-and-dont-give-in-51.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cm4LEgNJbWciZWh0dHBzOi8vdGhlY2hpdmUuZmlsZXMud29yZHByZXNzLmNvbS8yMDE1LzAyL2lmLXlvdXItZXgtdGV4dHMteW91LXN0YXktc3Ryb25nLWFuZC1kb250LWdpdmUtaW4tNTEuanBnDA", "is_valid": true},
    {"category": [], "rating": null, "title": "Family Feud contestant might have answered a little too honestly", "url": "https://thechive.files.wordpress.com/2015/02/screen-shot-2015-02-15-at-4-32-31-pm.png", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cl4LEgNJbWciVWh0dHBzOi8vdGhlY2hpdmUuZmlsZXMud29yZHByZXNzLmNvbS8yMDE1LzAyL3NjcmVlbi1zaG90LTIwMTUtMDItMTUtYXQtNC0zMi0zMS1wbS5wbmcM", "is_valid": true},
    {"category": [], "rating": null, "title": "These people take lazy to a new level", "url": "https://thechive.files.wordpress.com/2015/02/these-people-take-lazy-to-a-new-level-23-photos-22-e14235855147421.jpg", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cnwLEgNJbWcic2h0dHBzOi8vdGhlY2hpdmUuZmlsZXMud29yZHByZXNzLmNvbS8yMDE1LzAyL3RoZXNlLXBlb3BsZS10YWtlLWxhenktdG8tYS1uZXctbGV2ZWwtMjMtcGhvdG9zLTIyLWUxNDIzNTg1NTE0NzQyMS5qcGcM", "is_valid": true}
    ],
    "creator": {
    "name": "Ben",
    "img": "http://0.gravatar.com/avatar/67403a19b8ff2589cad1002324aaad88?s=50&d=http%3A%2F%2F0.gravatar.com%2Favatar%2Fad516503a11cd5ca435acc9bb6523536%3Fs%3D50&r=X"
    },
    "link": "http://thechive.com/2015/02/15/in-case-you-missed-them-check-out-the-top-posts-of-the-week-10-photos-2/",
    "date": "Sun, 15 Feb 2015 22:48:43 +0000",
    "guid": "http://thechive.com/?p=932925",
    "tags": ["Awesome", "Funny", "of", "posts", "the", "top", "week"]
    },
    {
    "author": null,
    "title": "Have the ultimate NBA slumber party in the Bulls owner\u2019s suite (8 HQ Photos)",
    "media": [
    {"category": [], "rating": null, "title": "The website committed to getting you into someone\u2019s house with the lowest chance of getting you murdered, Airbnb, is offering you a chance to win a \u201cNight at Home of the Chicago Bulls.\u201d If you enter before April 3rd (don\u2019t let March Madness distract you), you could cash in on a bedroom with a hell of a view; floor-to-ceiling of the court at the United Center. The winner will get courtside seats provided by the host, Scottie Pippen, who will take you on a post-game tour that culminates with a shoot-around on court. \n\nAs if that wasn\u2019t enough, you\u2019ll then watch a movie together on the Jumbotron and then retire to the insane suite where you will be the only person in history to stay the night in the United Center\u2019s suite. If you\u2019re a Bulls fan, check out the amazing digs and enter here. Just don\u2019t try to score on Scottie.", "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-1.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy0xLmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true},
    {"category": [], "rating": null, "title": null, "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-2.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy0yLmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true},
    {"category": [], "rating": null, "title": null, "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-4.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy00LmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true},
    {"category": [], "rating": null, "title": null, "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-3.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy0zLmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true},
    {"category": [], "rating": null, "title": null, "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-5.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy01LmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true},
    {"category": [], "rating": null, "title": null, "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-6.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy02LmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true},
    {"category": [], "rating": null, "title": null, "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-7.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy03LmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true},
    {"category": [], "rating": null, "title": null, "url": "https://thechive.files.wordpress.com/2015/03/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos-8.jpg?quality=94&strip=all", "urlsafe": "ahNzfmNydWNpYWwtYWxwaGEtNzA2cpsBCxIDSW1nIpEBaHR0cHM6Ly90aGVjaGl2ZS5maWxlcy53b3JkcHJlc3MuY29tLzIwMTUvMDMvaGF2ZS10aGUtdWx0aW1hdGUtbmJhLXNsdW1iZXItcGFydHktaW4tdGhlLWJ1bGxzLW93bmVycy1zdWl0ZS04LWhxLXBob3Rvcy04LmpwZz9xdWFsaXR5PTk0JnN0cmlwPWFsbAw", "is_valid": true}
    ],
    "creator": {
    "name": "Cameron",
    "img": "http://0.gravatar.com/avatar/34d7c66f1b69eda89120c190f7d61a79?s=50&d=http%3A%2F%2F0.gravatar.com%2Favatar%2Fad516503a11cd5ca435acc9bb6523536%3Fs%3D50&r=X"
    },
    "link": "http://thechive.com/2015/03/30/have-the-ultimate-nba-slumber-party-in-the-bulls-owners-suite-8-hq-photos/",
    "date": "Mon, 30 Mar 2015 18:30:00 +0000",
    "guid": "http://thechive.com/?p=966538",
    "tags": ["Awesome", "High-Res", "Sports", "contest", "luxury", "money", "NBA", "win"]}]}`)
}
