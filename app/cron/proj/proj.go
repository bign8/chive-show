package proj

import (
	"encoding/xml"
	"log"
	"sync"

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

// TODO: improve pulling performance (cache number of xml in stage_1, fan out pulling)
func puller(c appengine.Context) <-chan []byte {
	out := make(chan []byte, 10000)

	go func() {
		defer close(out)
		q := datastore.NewQuery(crawler.XML)
		t := q.Run(c)
		for {
			var s crawler.Store
			_, err := t.Next(&s)
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

func flatten(c appengine.Context, in <-chan []byte) <-chan Item {
	const WORKERS = 100
	out := make(chan Item, 10000)
	var wg sync.WaitGroup
	wg.Add(WORKERS)
	for i := 0; i < WORKERS; i++ {
		go func(idx int) {
			flattenWorker(c, in, out, idx)
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func flattenWorker(c appengine.Context, in <-chan []byte, out chan<- Item, idx int) {
	var xmlPage XMLPage
	var imgs []string

	for data := range in {
		if err := xml.Unmarshal(data, &xmlPage); err != nil {
			c.Errorf("Flatten %d: %v", idx, err)
			continue
		}
		for _, item := range xmlPage.Items {
			imgs = make([]string, len(item.Imgs))
			for i, img := range item.Imgs {
				imgs[i] = img.URL
			}

			out <- Item{
				GUID: item.GUID,
				Tags: item.Tags,
				Imgs: imgs,
			}
		}
	}
}

func doMagic() {
	start := make(chan interface{}, 10)
	out := chain.FanOut(10, 10, start, worker)
	go func() {
		for o := range out {
			log.Printf("Something: %v", o)
		}
	}()
	start <- 1
	start <- 2
	start <- 3
}

func worker(in <-chan interface{}, out chan<- interface{}, idx int) {
	var bytes []byte
	for x := range in {
		bytes = x.([]byte)
		out <- bytes
	}
}
