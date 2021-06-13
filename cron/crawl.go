package cron

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"

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
	var feed struct {
		Items []models.Post `xml:"channel>item"`
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
		}
	}

	// Cleanup Response
	// TODO: mine pages in parallel (worker pool?)
	for idx := range feed.Items {
		post := &feed.Items[idx]
		if err = mine(ctx, log.Default(), post); err != nil {
			log.Printf("Unable to mine page for details: %s", post.Link)
			return found, nil, err
		}
		for i, img := range post.Media {
			post.Media[i].URL = stripQuery(img.URL)
		}

		// Find the "author" image and remove it from the "media" set
		var found bool
		for i, media := range post.Media {
			if media.Category == "author" {
				found = true
				post.MugShot = media.URL
				post.Media = append(post.Media[:i], post.Media[i+1:]...)
				break
			}
		}
		if !found {
			log.Printf("Unable to find author for: %#v", post)
			post.MugShot = post.Media[0].URL
			post.Media = post.Media[1:]
		}
	}
	return found, feed.Items, nil
}
