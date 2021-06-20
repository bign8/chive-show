package cron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"go.opencensus.io/trace"

	"github.com/bign8/chive-show/appengine"
	"github.com/bign8/chive-show/models"
)

func MineHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.FormValue("id")
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			appengine.Error(r.Context(), "cron(mine): Failed to parse id(%s): %v", sid, err)
			http.Error(w, "unable to parse id", http.StatusExpectationFailed)
			return
		}

		// post, err := store.Get(r.Context(), id)
		// if err != nil {
		// 	appengine.Error(r.Context(), "cron(mine): Failed to load post(%d): %v", id, err)
		// 	http.Error(w, "failed to load post", http.StatusNotFound)
		// 	return
		// }

		// Mine the post data!
		post, err := mineFull(r.Context(), id)
		if err != nil {
			appengine.Error(r.Context(), "cron(mine): Failed to mine post(%d): %v", id, err)
			http.Error(w, "failed to mine post", http.StatusInternalServerError)
			return
		}

		// TODO: allow a dry run?
		err = store.Put(r.Context(), post)
		if err != nil {
			appengine.Error(r.Context(), "cron(mine): Failed to store post(%d): %v", id, err)
			http.Error(w, "failed to store post", http.StatusFailedDependency)
			return
		}

		// Print meaningful output to user
		enc := json.NewEncoder(w)
		enc.SetIndent(``, ` `)
		enc.SetEscapeHTML(false)
		if err = enc.Encode(post); err != nil {
			appengine.Error(r.Context(), "cron(mine): Failed to encode post(%d): %v", id, err)
		}
	}
}

// NOTE: there is an alternate link in the header of posts that points to a JSON representation of the post
// ex: https://thechive.com/?p=3701780 => https://thechive.com/wp-json/wp/v2/posts/3701780
// UPDATE: the JSON media links (located at ._links["wp:attachment"][0].href) are NOT in order! still need to mine page to fetch order
// Update: media links (https://thechive.com/wp-json/wp/v2/media?parent=3701780) DO contain image size .[].media_details.(height|width)
// Update: media links referencing "FULL" for gifs are actually GIFs!
func mine(ctx context.Context, post *models.Post) error {
	// if post.Version == nil && len(post.Media) > 0 {
	// 	post.Thumb = post.Media[0].URL
	// }
	// post.Version = &latestVersion
	// post.Media = nil // idempotent (reset previous runs data)
	body, err := fetch(ctx, post.Link)
	if err != nil {
		appengine.Error(ctx, "mine(%s): unable to fetch: %s", post.Link, err)
		return nil
	}
	dom, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	appengine.Info(ctx, "Loaded %d bytes from %s", len(dom), post.Link)
	if err = body.Close(); err != nil {
		return err
	}
	doc := soup.HTMLParse(string(dom))

	// Pages embed single banner image
	figure := doc.Find("figure")
	if figure.Error != nil {
		appengine.Warning(ctx, "unable to load figure %s: %v", post.Link, figure.Error)
		return nil
	}
	first, err := strconv.ParseInt(figure.Attrs()["data-attachment-id"], 10, 64)
	if err != nil {
		appengine.Warning(ctx, "Unable to load first figure ID: %v", err)
		return nil
	}
	obj := figure.Find("img")
	if obj.Error != nil {
		obj = figure.Find("source")
	}
	if obj.Error != nil {
		appengine.Warning(ctx, "uanble to load banner content %s: %v", post.Link, obj.Error)
	} else {
		media := models.Media{
			ID:  first,
			URL: stripQuery(obj.Attrs()["src"]),
		}

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
		appengine.Warning(ctx, "Unable to find script logic in %q %v", post.Link, js.Error)
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
			ID   int64  `json:"id"`
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
				appengine.Warning(ctx, "got nil HTML in item %d on %s", i, post.Link)
			}
			continue
		}
		ele := soup.HTMLParse(obj.HTML)
		if ele.Error != nil {
			appengine.Warning(ctx, "unable to parse HTML of %d on %s", i, post.Link)
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
			appengine.Warning(ctx, "No media found in item %d on %s", i, post.Link)
			continue
		}
		for _, img := range imgs {
			media := models.Media{
				ID:  obj.ID,
				URL: stripQuery(img.Attrs()["src"]),
			}

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

// NOTE: there is an alternate link in the header of posts that points to a JSON representation of the post
// ex: https://thechive.com/?p=3701780 => https://thechive.com/wp-json/wp/v2/posts/3701780
// UPDATE: the JSON media links (located at ._links["wp:attachment"][0].href) are NOT in order! still need to mine page to fetch order
// Update: media links (https://thechive.com/wp-json/wp/v2/media?parent=3701780) DO contain image size .[].media_details.(height|width)
// Update: media links referencing "FULL" for gifs are actually GIFs!
// Future: until we can load the post order from the API, see if we can just load the HTML page and the attachments API for media dimensions :shrug:
func mineFull(ctx context.Context, id int64) (*models.Post, error) {
	result := &models.Post{
		ID:      id,
		Version: 1,
	}
	var (
		createrID int
		meta      = make(map[int64]models.Media)
	)

	// Read the JSON information about the page
	// found this guy by reviewing the source of a chive post: link rel="alternate" type="application/json"
	link := fmt.Sprintf("https://thechive.com/wp-json/wp/v2/posts/%d", id)
	appengine.Info(ctx, "Link: %s", link)
	err := fetchParse(ctx, link, func(body io.Reader) error {
		var postInfo struct {
			Date   string     `json:"date_gmt"`
			Link   string     `json:"link"`
			Title  renderable `json:"title"`
			Author int        `json:"author"`
			Thumb  string     `json:"jetpack_featured_media_url"`
			Links  map[string][]struct {
				Href string `json:"href"`
			} `json:"_links"`
		}
		err := json.NewDecoder(body).Decode(&postInfo)
		if err != nil {
			appengine.Error(ctx, "postInfo json decode failed: %v", err)
			return err
		}
		// info.Printf("postInfo: %v", postInfo)
		result.Thumb = stripQuery(postInfo.Thumb)
		result.Title = postInfo.Title.Rendered
		result.Link = postInfo.Link
		if result.Date, err = time.Parse(`2006-01-02T15:04:05`, postInfo.Date); err != nil {
			appengine.Error(ctx, "Unable to parse date %q: %v", postInfo.Date, err)
			return err
		}
		createrID = postInfo.Author
		return nil
	})
	if err != nil {
		return result, err
	}

	// Load the metadata for all the media (NOTE: this only loads 10 pieces of media at a time, need to load multiple times)
	for i := int64(1); i < 20 && err == nil; i++ {
		err = fetchMediaPage(ctx, id, i, meta)
	}
	if badFetch, ok := err.(errBadFetch); ok {
		if badFetch.code == 400 {
			err = nil // API appears to return 400s when fetching more pages that possible
		}
	}
	if err != nil {
		return result, err
	}

	// Load Post Creator
	createrLink := fmt.Sprintf("https://thechive.com/wp-json/wp/v2/users/%d", createrID)
	appengine.Info(ctx, "Creator: %s", createrLink)
	err = fetchParse(ctx, createrLink, func(body io.Reader) error {
		var authorData struct {
			Name string            `json:"name"`
			URLs map[string]string `json:"avatar_urls"`
		}
		if err = json.NewDecoder(body).Decode(&authorData); err != nil {
			appengine.Error(ctx, "authorData json decode failed: %v", err)
			return err
		}
		result.Creator = authorData.Name
		for _, link := range authorData.URLs {
			result.MugShot = stripQuery(link)
			return nil
		}
		return errors.New("no author avatar found")
	})
	if err != nil {
		return result, err
	}

	// Load Post Tags
	tagsLink := fmt.Sprintf("https://thechive.com/wp-json/wp/v2/categories?post=%d", id)
	appengine.Info(ctx, "Tags: %s", tagsLink)
	err = fetchParse(ctx, tagsLink, func(body io.Reader) error {
		var tagData []struct {
			Name string `json:"name"`
		}
		if err = json.NewDecoder(body).Decode(&tagData); err != nil {
			appengine.Error(ctx, "tagData json decode failed: %v", err)
			return err
		}
		for _, tag := range tagData {
			result.Tags = append(result.Tags, tag.Name)
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	// Reload a majority of the post to fetch the correct asset order
	err = mine(ctx, result)
	if err != nil {
		return result, err
	}

	// Assign the media back to the API's result after fetching HTML's content
	missed := make(map[int64]bool)
	for i, media := range result.Media {
		m, ok := meta[media.ID]
		if !ok {
			missed[media.ID] = true
			continue
		}
		result.Media[i] = m
		delete(meta, media.ID)
	}

	// Print error for missing media metadata
	if len(missed) > 0 {
		appengine.Warning(ctx, "Unable to find media %v in meta %v", missed, meta)
	}
	if len(meta) > 0 {
		appengine.Warning(ctx, "Leftover meta from JSON endpoint: %v", meta)
	}

	return result, nil
}

type renderable struct {
	Rendered string `json:"rendered"`
}

func fetchParse(ctx context.Context, url string, fn func(io.Reader) error) error {
	body, err := fetch(ctx, url)
	if err != nil {
		return err
	}
	defer body.Close()
	return fn(body)
}

func fetchMediaPage(ctx context.Context, id, page int64, meta map[int64]models.Media) error {
	mediaLink := fmt.Sprintf("https://thechive.com/wp-json/wp/v2/media?parent=%d&page=%d", id, page)
	appengine.Info(ctx, "Media: %s", mediaLink)
	return fetchParse(ctx, mediaLink, func(body io.Reader) error {
		var mediaData []struct {
			ID      int64      `json:"id"`
			GUID    renderable `json:"guid"`
			Caption renderable `json:"caption"`
			Details struct {
				Height int `json:"height"`
				Width  int `json:"width"`
			} `json:"media_details"`
		}
		if err := json.NewDecoder(body).Decode(&mediaData); err != nil {
			appengine.Error(ctx, "mediaData json decode failed: %v", err)
			return err
		}
		// info.Printf("mediaData: %v", mediaData)
		for _, media := range mediaData {
			meta[media.ID] = models.Media{
				ID:      media.ID,
				URL:     media.GUID.Rendered,
				Caption: strings.TrimSpace(media.Caption.Rendered),
				Height:  media.Details.Height,
				Width:   media.Details.Width,
			}
		}
		return nil
	})
}

func mineMedia(rctx context.Context, post *models.Post) error {
	ctx, span := trace.StartSpan(rctx, "cron.mineMedia")
	defer span.End()
	span.AddAttributes(trace.Int64Attribute(`id`, post.ID))

	if err := mine(ctx, post); err != nil {
		return err
	}

	meta := make(map[int64]models.Media)
	limit := int64(math.Ceil(float64(len(post.Media)) / 10)) // paginated to 10 pages per thing
	for i := int64(1); i <= limit; i++ {
		if err := fetchMediaPage(ctx, post.ID, i, meta); err != nil {
			return err
		}
	}

	// Assign the media back to the API's result after fetching HTML's content
	missed := make(map[int64]bool)
	for i, media := range post.Media {
		m, ok := meta[media.ID]
		if !ok {
			missed[media.ID] = true
			continue
		}
		post.Media[i] = m
		delete(meta, media.ID)
	}

	// Print error for missing media metadata
	if len(missed) > 0 {
		appengine.Warning(ctx, "Unable to find media %v in meta %v", missed, meta)
	}
	if len(meta) > 0 {
		appengine.Warning(ctx, "Leftover meta from JSON endpoint: %v", meta)
	}

	return nil
}
