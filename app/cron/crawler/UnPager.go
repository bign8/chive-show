package crawler

import (
	"encoding/xml"

	"appengine"
)

// UnPager process pages of posts to individual posts
func UnPager(c appengine.Context, pages <-chan string) <-chan Data {
	res := make(chan Data)

	// TODO: spin up as many unpages as desired
	go runUnPager(c, pages, res)
	return res
}

func runUnPager(c appengine.Context, in <-chan string, out chan<- Data) {
	defer close(out)

	var miner struct {
		Item []struct {
			KEY string `xml:"guid"`
			XML string `xml:",innerxml"`
		} `xml:"channel>item"`
	}

	for page := range in {
		c.Infof("UnPager: Retrieved Page")

		if err := xml.Unmarshal([]byte(page), &miner); err != nil {
			c.Errorf("UnPager: Error %s", err)
		}

		for _, post := range miner.Item {
			c.Infof("UnPager: Found Post %s", post.KEY)
			out <- Data{
				KEY: post.KEY,
				XML: post.XML,
			}
		}
	}
}
