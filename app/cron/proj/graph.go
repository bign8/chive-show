package proj

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"appengine"

	"github.com/bign8/chive-show/app/cron/proj/graph"
	"github.com/bign8/chive-show/app/helpers/sharder"
)

// TestShard to delete
func TestShard(c appengine.Context, w http.ResponseWriter, r *http.Request) {

	data := []byte(strings.Repeat("01234567890123456789", 1e6))

	// Writing
	start := time.Now()
	err := sharder.Writer(c, "test", data)
	if err != nil {
		c.Errorf("Writer Error: %s", err)
		return
	}
	c.Infof("Write took: %v", time.Since(start))

	// Reading
	start = time.Now()
	read, err := sharder.Reader(c, "test")
	if err != nil {
		c.Errorf("Reader Error: %s", err)
		return
	}
	c.Infof("Data Length: %d; isSame: %v", len(read), bytes.Equal(read, data))
	c.Infof("Read took: %v", time.Since(start))
}

// Graph processes all posts in attempt to create a graph
func Graph(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var item Item
	var post, ntag, nimg *graph.Node

	idx := 0
	timeout := time.After(time.Second)
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

		// This is a DEBUG only operation
		select {
		case <-timeout:
			c.Infof("Index: %d; Duration: %v", idx, time.Since(start))
			timeout = time.After(time.Second)
		default:
		}
		idx++
	}
	c.Infof("End Loop: %d; Duration: %v", idx, time.Since(start))

	// Write result
	bits, err := g.Bytes()
	if err != nil {
		c.Errorf("Error in Graph.Bytes: %v", err)
	}
	c.Infof("End Serialization: Len(%d); Duration: %v", len(bits), time.Since(start))

	// Storage
	if err := sharder.Writer(c, "graph", bits); err != nil {
		c.Errorf("Writer Error: %s", err)
		return
	}
	c.Infof("Write Complete; Duration: %v", time.Since(start))

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
	// INFO: Nodes (IMG): 928728
	// INFO: Nodes (TAG): 244212
	// INFO: Nodes (POST): 40920
	// INFO: Nodes (ALL): 1213860
	// INFO: Time took: 31.310686059s

	// w/dupes w/o invalid Tags
	// INFO: Nodes (IMG): 928728
	// INFO: Nodes (TAG): 237122
	// INFO: Nodes (POST): 40920
	// INFO: Nodes (ALL): 1206770
	// INFO: Time took: 31.850210891s

	// w/o dupes w/o invalid Tags
	// INFO: Nodes (IMG): 886831
	// INFO: Nodes (POST): 40920
	// INFO: Nodes (TAG): 18221
	// INFO: Nodes (ALL): 945972
	// INFO: Time took: 32.651739532s

	// TODO: write to sharded datastore entity

	c.Infof("Time took: %v", time.Since(start))
}
