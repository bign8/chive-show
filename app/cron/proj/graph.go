package proj

import (
	"net/http"
	"time"

	"appengine"

	"github.com/bign8/chive-show/app/cron/proj/graph"
)

// Graph processes all posts in attempt to create a graph
func Graph(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var item Item
	var post, ntag, nimg *graph.Node

	idx := 0

	g := graph.New(false)
	for idk := range getItems(c) {
		item = idk.(Item)
		post = g.Add(item.GUID, graph.NodeType_POST, 0)

		for _, tag := range validTags(item.Tags) {
			ntag = g.Add(tag, graph.NodeType_TAG, 0)
			g.Connect(post, ntag, 0)
		}

		for _, img := range item.Imgs {
			nimg = g.Add(img, graph.NodeType_IMG, 0)
			g.Connect(post, nimg, 0)
		}

		// This is a SLOW/DEBUG only operation
		if idx%2000 == 0 {
			c.Infof("Current Duration (%v)", time.Since(start))
		}
		idx++
	}

	// Write result
	bits, err := g.Bytes()
	if err != nil {
		c.Errorf("Error in Graph.Bytes: %v", err)
	}
	w.Write(bits)

	// Count types of nodes
	binCtr := make(map[graph.NodeType]uint64)
	for _, node := range g.Nodes() {
		binCtr[*node.Type]++
	}

	// Log out types of nodes
	total := uint64(0)
	for key, value := range binCtr {
		c.Infof("Nodes (%s): %d", key, value)
		total += value
	}
	c.Infof("Nodes (ALL): %d", total)

	// w/dupes w/invalid tags
	// 2015/11/15 20:52:26 INFO: Nodes (IMG): 928728
	// 2015/11/15 20:52:26 INFO: Nodes (TAG): 244212
	// 2015/11/15 20:52:26 INFO: Nodes (POST): 40920
	// 2015/11/15 20:52:26 INFO: Time took: 31.310686059s

	// w/dupes w/o invalid Tags
	// 2015/11/15 21:03:06 INFO: Nodes (IMG): 928728
	// 2015/11/15 21:03:06 INFO: Nodes (TAG): 237122
	// 2015/11/15 21:03:06 INFO: Nodes (POST): 40920
	// 2015/11/15 21:03:06 INFO: Time took: 31.850210891s

	// w/o dupes w/o invalid Tags
	// 2015/11/15 21:06:18 INFO: Nodes (IMG): 886831
	// 2015/11/15 21:06:18 INFO: Nodes (POST): 40920
	// 2015/11/15 21:06:18 INFO: Nodes (TAG): 18221
	// 2015/11/15 21:06:18 INFO: Time took: 32.651739532s

	c.Infof("Time took: %v", time.Since(start))
}
