package cron

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	// err := mine(w, "https://thechive.com/2021/05/20/daily-afternoon-randomness-49-photos-1287/")
	err := mine(w, "https://thechive.com/2021/05/28/daily-morning-awesomeness-38-photos-151/")
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
		log.Printf("mine(%s): unable to fetch: %s", link, err)
		return nil
	}
	doc := soup.HTMLParse(res)

	for _, figure := range doc.FindAll("figure") {
		img := stripQuery(figure.Find("img").Attrs()["src"])
		fmt.Fprintf(w, "RAW HTML : %s\n\n", img)
	}

	// parse CHIVE_GALLERY_ITEMS from script id='chive-theme-js-js-extra' into JSON
	// TODO: use match the image prefix? "https:\/\/thechive.com\/wp-content\/uploads\/" in the HTML and parse to closing "
	src := doc.Find("script", "id", "chive-theme-js-js-extra").FullText()
	idx := strings.IndexByte(src, '{')
	if idx < 0 {
		return errors.New("unable to find opening brace")
	}
	src = src[idx:]
	idx = strings.LastIndexByte(src, '}')
	if idx < 0 {
		return errors.New("unable to find closing brace")
	}
	src = src[:idx+1]
	var what struct {
		Items []struct {
			HTML *string `json:"html,omitempty"`
		} `json:"items"`
	}
	err = json.Unmarshal([]byte(src), &what)
	if err != nil {
		panic(err)
	}

	// Parse HTML attributes of JSON to get images
	for i, obj := range what.Items {
		if obj.HTML == nil {
			log.Printf("no HTML found in fragment %d (embedded in post?)", i)
			continue
		}
		ele := soup.HTMLParse(*obj.HTML)
		if ele.Error != nil {
			log.Printf("unable to parse post fragment %d", i)
			continue
		}
		imgs := ele.FindAll("img")
		if len(imgs) == 0 {
			log.Printf("no images found in fragment %d (video?)", i)
			continue
		}
		for _, img := range imgs {
			img := stripQuery(img.Attrs()["src"])
			fmt.Fprintf(w, "JSON+HTML: %s\n\n", img)
		}
	}
	return nil
}
