package proj

import (
	"net/http"
	"time"

	"appengine"
)

// Graph processes all posts in attempt to create a graph
func Graph(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// pages := puller(c)
	// dirtyTags := getNod(c, pages, 100)
	// tags := cleaner(dirtyTags)
	//
	// found := map[string]int64{}
	// for tag := range tags {
	// 	found[tag]++
	// }
	//
	// for key, value := range found {
	// 	fmt.Fprintf(w, "%s,%d\n", key, value)
	// }

	c.Infof("Time took: %v", time.Since(start))
}
