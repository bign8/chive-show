package crawler

import "appengine"

// UnPager process pages of posts to individual posts
func UnPager(c appengine.Context, pages <-chan string) <-chan string {
	res := make(chan string)
	go runUnPager(c, pages, res)
	return res
}

func runUnPager(c appengine.Context, in <-chan string, out chan<- string) {
	defer close(out)

	for page := range in {
		c.Infof("Retrieved Page %s", page)

		// TODO: decompress page
	}
}
