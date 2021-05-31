package cron

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/bign8/chive-show/models"
)

func sample(w http.ResponseWriter, r *http.Request) {

	// last page the time of testing: 6305
	// multi(6305): 2008/12/12
	//
	// page, _ := strconv.Atoi(r.URL.Query().Get("page"))

	start := time.Now()
	// posts, err := sampled(r.Context(), page)
	// if err != nil {
	// 	panic(err)
	// }

	err := mine(w, "https://thechive.com/2021/05/20/daily-afternoon-randomness-49-photos-1287/")
	if err != nil {
		panic(err)
	}

	// enc := json.NewEncoder(w)
	// enc.SetIndent("", " ")
	// enc.Encode(posts)
	log.Printf("Parsing took %s", time.Since(start))
}

func sampled(ctx context.Context, idx int) ([]models.Post, error) {
	url := pageURL(idx)

	// Get Response
	log.Printf("INFO: Parsing index %v (%v)", idx, url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, ErrFeedParse404
		}
		return nil, fmt.Errorf("feed parcing recieved a %d Status Code", resp.StatusCode)
	}

	// Decode Response
	decoder := xml.NewDecoder(resp.Body)
	var feed struct {
		Items []models.Post `xml:"channel>item"`
	}
	if decoder.Decode(&feed) != nil {
		return nil, err
	}

	// Cleanup Response
	for idx := range feed.Items {
		post := &feed.Items[idx]
		for i, img := range post.Media {
			post.Media[i].URL = stripQuery(img.URL)
		}
		// NOTE: this can be the case for 3 photos as well
		if len(post.Media) == 2 && post.Media[1].Category == "author" {
			post.MugShot = post.Media[1].URL
			post.Media = post.Media[:1]
			// TODO: IDK when, but new posts don't include all the images, chase that down!
		} else {
			post.MugShot = post.Media[0].URL
			post.Media = post.Media[1:]
		}
	}
	return feed.Items, err
}

func mine(w http.ResponseWriter, link string) error {
	res, err := soup.Get(link)
	if err != nil {
		panic(err)
	}
	doc := soup.HTMLParse(res)
	for _, figure := range doc.FindAll("figure") {
		fmt.Fprintf(w, "Figure: %s<br/>\n", figure.Find("img").FullText())
	}
	// log.Printf("Got soup: %v", doc)
	return nil
}
