package proj

import (
	"encoding/xml"

	"appengine"
	"appengine/datastore"

	"github.com/bign8/chive-show/app/cron/chain"
	"github.com/bign8/chive-show/app/cron/crawler"
)

// XMLPage xml processor for a page
type XMLPage struct {
	Items []struct {
		GUID string   `xml:"guid"`
		Tags []string `xml:"category"`
		Imgs []struct {
			URL string `xml:"url,attr"`
		} `xml:"content"`
	} `xml:"channel>item"`
}

// Item is a post item
type Item struct {
	GUID string
	Tags []string
	Imgs []string
}

func getItems(c appengine.Context) <-chan interface{} {
	pages := puller(c)
	return chain.FanOut(50, 10000, pages, flatten(c))
}

func puller(c appengine.Context) <-chan interface{} {
	out := make(chan interface{}, 10000)

	// TODO: improve pulling performance (cache number of xml in stage_1, fan out pulling)
	go func() {
		defer close(out)
		q := datastore.NewQuery(crawler.XML)
		iterator := q.Run(c)
		for {
			var s crawler.Store
			_, err := iterator.Next(&s)
			if err == datastore.Done {
				break // No further entities match the query.
			}
			if err != nil {
				c.Errorf("fetching next Person: %v", err)
				break
			}
			out <- s.XML
		}
	}()
	return out
}

func flatten(c appengine.Context) chain.Worker {
	return func(obj interface{}, out chan<- interface{}, idx int) {
		var xmlPage XMLPage
		var imgs []string

		// Parse the XML of an object
		if err := xml.Unmarshal(obj.([]byte), &xmlPage); err != nil {
			c.Errorf("Flatten %d: %v", idx, err)
			return
		}

		// Process items in a particular page
		for _, item := range xmlPage.Items {
			imgs = make([]string, len(item.Imgs))
			for i, img := range item.Imgs {
				imgs[i] = img.URL
			}
			out <- Item{item.GUID, item.Tags, imgs}
		}
	}
}
