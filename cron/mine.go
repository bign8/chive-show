package cron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
	"github.com/bign8/chive-show/models"
)

func MineHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.FormValue("id")
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			log.Printf("cron(mine): Failed to parse id(%s): %v", sid, err)
			http.Error(w, "unable to parse id", http.StatusExpectationFailed)
			return
		}

		post, err := store.Get(r.Context(), id)
		if err != nil {
			log.Printf("cron(mine): Failed to load post(%d): %v", id, err)
			http.Error(w, "failed to load post", http.StatusNotFound)
			return
		}

		// Create logger to user can see whats going on too!
		info := log.New(
			io.MultiWriter(w, log.Default().Writer()),
			fmt.Sprintf("mine(%d): ", id),
			0,
		)

		// Mine the post data!
		err = mine(r.Context(), info, post)
		if err != nil {
			log.Printf("cron(mine): Failed to mine post(%d): %v", id, err)
			http.Error(w, "failed to mine post", http.StatusInternalServerError)
			return
		}

		// // TODO: allow a dry run?
		// err = store.Put(r.Context(), post)
		// if err != nil {
		// 	log.Printf("cron(mine): Failed to store post(%d): %v", id, err)
		// 	http.Error(w, "failed to store post", http.StatusFailedDependency)
		// 	return
		// }

		// Print meaningful output to user
		enc := json.NewEncoder(w)
		enc.SetIndent(``, ` `)
		enc.SetEscapeHTML(false)
		if err = enc.Encode(post); err != nil {
			log.Printf("cron(mine): Failed to encode post(%d): %v", id, err)
		}
	}
}

// NOTE: there is an alternate link in the header of posts that points to a JSON representation of the post
// ex: https://thechive.com/?p=3701780 => https://thechive.com/wp-json/wp/v2/posts/3701780
// UPDATE: the JSON media links (located at ._links["wp:attachment"][0].href) are NOT in order! still need to mine page to fetch order
// Update: media links (https://thechive.com/wp-json/wp/v2/media?parent=3701780) DO contain image size .[].media_details.(height|width)
// Update: media links referencing "FULL" for gifs are actually GIFs!
func mine(ctx context.Context, info *log.Logger, post *models.Post) error {
	// if post.Version == nil && len(post.Media) > 0 {
	// 	post.Thumb = post.Media[0].URL
	// }
	// post.Version = &latestVersion
	// post.Media = nil // idempotent (reset previous runs data)
	body, err := fetch(ctx, post.Link)
	if err != nil {
		log.Printf("mine(%s): unable to fetch: %s", post.Link, err)
		return nil
	}
	dom, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	info.Printf("Loaded %d bytes from %s", len(dom), post.Link)
	if err = body.Close(); err != nil {
		return err
	}
	doc := soup.HTMLParse(string(dom))

	// Pages embed single banner image
	figure := doc.Find("figure")
	if figure.Error != nil {
		log.Printf("WARNING: unable to load figure %s: %v", post.Link, figure.Error)
		return nil
	}
	obj := figure.Find("img")
	if obj.Error != nil {
		obj = figure.Find("source")
	}
	if obj.Error != nil {
		log.Printf("WARNING: uanble to load banner content %s: %v", post.Link, obj.Error)
	} else {
		media := models.Media{URL: obj.Attrs()["src"]}

		// Attempt to scrape captions as well
		caption := figure.Find("figcaption", "class", "gallery-caption")
		if caption.Error == nil {
			for _, ele := range caption.Children() {
				media.Caption += ele.HTML()
			}
			media.Caption = strings.TrimSpace(media.Caption)
		}

		post.Media = append(post.Media, media)
	}

	// parse CHIVE_GALLERY_ITEMS from script id='chive-theme-js-js-extra' into JSON
	// TODO: use match the image prefix? "https:\/\/thechive.com\/wp-content\/uploads\/" in the HTML and parse to closing "
	js := doc.Find("script", "id", "chive-theme-js-js-extra")
	if js.Error != nil {
		log.Printf("WARNING: Unable to find script logic in %q %v", post.Link, js.Error)
		return nil
	}
	src := js.FullText()
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
			HTML string `json:"html"`
			Type string `json:"type"`
		} `json:"items"`
	}
	err = json.Unmarshal([]byte(src), &what)
	if err != nil {
		panic(err)
	}

	// Parse HTML attributes of JSON to get images
	for i, obj := range what.Items {
		if obj.HTML == "" {
			// first entry allways appears to be empty as the first post is embedded in page content
			// that said, this warning is here in case something changes in the future
			if i != 0 {
				log.Printf("WARNING: got nil HTML in item %d on %s", i, post.Link)
			}
			continue
		}
		ele := soup.HTMLParse(obj.HTML)
		if ele.Error != nil {
			log.Printf("WARNING: unable to parse HTML of %d on %s", i, post.Link)
			continue
		}
		var imgs []soup.Root
		switch obj.Type {
		case "gif":
			imgs = ele.FindAll("source")
		default:
			imgs = ele.FindAll("img")
		}
		if len(imgs) == 0 {
			log.Printf("WARNING: No media found in item %d on %s", i, post.Link)
			continue
		}
		for _, img := range imgs {
			media := models.Media{URL: stripQuery(img.Attrs()["src"])}

			// Attempt to scrape captions as well
			caption := ele.Find("figcaption", "class", "gallery-caption")
			if caption.Error == nil {
				for _, ele := range caption.Children() {
					media.Caption += ele.HTML()
				}
				media.Caption = strings.TrimSpace(media.Caption)
			}

			post.Media = append(post.Media, media)
		}
	}
	return nil
}
