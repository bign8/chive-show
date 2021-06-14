package cron

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/bign8/chive-show/models"
)

func CrawlHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, posts, err := getAndParseFeed(r.Context(), store, 1)
		if err != nil {
			log.Printf("crawl: failed to load page: %v", err)
			http.Error(w, "failed to load feed", http.StatusFailedDependency)
			return
		}

		err = store.PutMulti(r.Context(), posts)
		if err != nil {
			log.Printf("crawl: failed to store %d posts: %v", len(posts), err)
			http.Error(w, "failed to store posts", http.StatusInternalServerError)
			return
		}

		log.Printf("Stored %d posts", len(posts))
		enc := json.NewEncoder(w)
		enc.SetIndent(``, ` `)
		enc.SetEscapeHTML(false)
		enc.Encode(posts)
	}
}

var clean = regexp.MustCompile(`\W\(([^\)]*)\)$`)

func getAndParseFeed(ctx context.Context, store models.Store, idx int) (found int, posts []models.Post, err error) {
	url := pageURL(idx)

	// Get Response
	log.Printf("INFO: Parsing index %v (%v)", idx, url)
	body, err := fetch(ctx, url)
	if err != nil {
		return 0, nil, err
	}
	defer body.Close()

	// Decode Response
	decoder := xml.NewDecoder(body)
	// FUTURE: use a real object here an implement xml.Unmarshaler (see below)
	var feed struct {
		Items []struct {
			ID      int64    `xml:"post-id"`
			Tags    []string `xml:"category"`
			Link    string   `xml:"link"`
			Date    string   `xml:"pubDate"`
			Title   string   `xml:"title"`
			Creator string   `xml:"creator"`
			Media   []struct {
				URL      string `xml:"url,attr"`
				Type     string `xml:"type,attr"`
				Title    string `xml:"-"` // the xml titles are worthless (especially now that the full content isn't present)
				Rating   string `xml:"rating"`
				Category string `xml:"category"`
			} `xml:"content"`
		} `xml:"channel>item"`
	}
	if err := decoder.Decode(&feed); err != nil {
		return 0, nil, err
	}

	// Remove undesired posts to reduce the amount of mined data and trash in the database
	//  - dopamine dump posts (they link out to i.thechive.com and host external user content)
	//  - duplicates!
	remove := func(i int) {
		feed.Items = append(feed.Items[:i], feed.Items[i+1:]...)
	}
	found = len(feed.Items)
	for i := len(feed.Items) - 1; i >= 0; i-- {
		post := feed.Items[i]
		if post.Link == "http://i.thechive.com/dopamine-dump" {
			// https://i.thechive.com/rest/uploads?queryType=dopamine-dump&offset=0
			log.Printf("INFO: Ignoring dopamine dump (TODO: separate miner for i.thechive.com/rest/uploads): %s", post.Link)
			remove(i)
		} else if has, err := store.Has(ctx, post.ID); has || err != nil {
			log.Printf("INFO: Found duplicate: %s %v", post.Link, err)
			remove(i)
		} else {
			// Remove singular videos
			for i, media := range post.Media {
				if media.Type == "video/mp4" {
					log.Printf("INFO: Found video post: ignoring for now: %s", post.Link)
					remove(i)
					break
				}
			}
		}

	}

	// Cleanup Response
	// TODO: mine pages in parallel (worker pool?)
	for _, xmlPost := range feed.Items {
		post := models.Post{
			ID:      xmlPost.ID,
			Tags:    xmlPost.Tags,
			Link:    xmlPost.Link,
			Title:   clean.ReplaceAllLiteralString(xmlPost.Title, ``),
			Creator: xmlPost.Creator,
		}

		// Convert the date to a real date
		if post.Date, err = time.Parse(time.RFC1123Z, xmlPost.Date); err != nil {
			return 0, nil, err
		}

		// Find the "author" image and remove it from the "media" set
		if len(xmlPost.Media) != 2 {
			log.Printf("Found more than 2 medias for: %#v", xmlPost)
		}
		for _, media := range xmlPost.Media {
			if media.Category == "author" {
				post.MugShot = stripQuery(media.URL)
			} else {
				post.Thumb = stripQuery(media.URL)
			}
		}

		if err = mineMedia(ctx, log.Default(), &post); err != nil {
			log.Printf("Unable to mine page for details: %s", post.Link)
			return found, nil, err
		}
		posts = append(posts, post)
	}
	return found, posts, nil
}

// // UnmarshalXML implements xml.Unmarshaler for custom unmarshaling logic : https://golang.org/pkg/encoding/xml/#Unmarshaler
// func (x *Post) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
// 	var post struct {
// 		ID      int64    `xml:"post-id"`
// 		Tags    []string `xml:"category"`
// 		Link    string   `xml:"link"`
// 		Date    string   `xml:"pubDate"`
// 		Title   string   `xml:"title"`
// 		Creator string   `xml:"creator"`
// 		Media   []Media  `xml:"content"`
// 	}
// 	if err = d.DecodeElement(&post, &start); err != nil {
// 		return err
// 	}
// 	x.ID = post.ID
// 	x.Tags = post.Tags
// 	x.Link = post.Link
// 	x.Title = clean.ReplaceAllLiteralString(post.Title, "")
// 	x.Creator = post.Creator
// 	x.Media = post.Media // TODO: do mugshot scrubbing here
// 	if x.Date, err = time.Parse(time.RFC1123Z, post.Date); err != nil {
// 		return err
// 	}
// 	return nil
// }
