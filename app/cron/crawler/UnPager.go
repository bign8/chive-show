package crawler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"sync"

	"appengine"
)

// UnPage unpage and flatten data from Crawling
func UnPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	pages := Puller(c, XML)
	posts := UnPager(c, pages, 10)

	vertexes, edges := Miner(c, posts, 30)

	vbatch := Batcher(c, vertexes, 100)
	ebatch := Batcher(c, edges, 100)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		Storage(c, vbatch, 10, "vertex")
		wg.Done()
	}()
	go func() {
		Storage(c, ebatch, 10, "edge")
		wg.Done()
	}()
	wg.Wait()

	fmt.Fprintf(w, "Done")
}

// UnPager process pages of posts to individual posts
func UnPager(c appengine.Context, pages <-chan string, workers int) <-chan Data {
	res := make(chan Data, 100000)

	// TODO: spin up as many unpages as desired
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(x int) {
			runUnPager(c, pages, res, x)
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		close(res)
	}()

	return res
}

func runUnPager(c appengine.Context, in <-chan string, out chan<- Data, idx int) {
	var miner struct {
		Item []struct {
			KEY string `xml:"guid"`
			XML string `xml:",innerxml"`
		} `xml:"channel>item"`
	}

	for page := range in {
		// c.Infof("UnPager %d: Retrieved Page", idx)

		if err := xml.Unmarshal([]byte(page), &miner); err != nil {
			c.Errorf("UnPager: Error %s", err)
		}

		for _, post := range miner.Item {
			// c.Infof("UnPager: Found Post %s", post.KEY)
			out <- Data{
				KEY: post.KEY,
				XML: post.XML,
			}
		}
	}
}
