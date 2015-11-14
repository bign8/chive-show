package crawler

import (
	"encoding/xml"
	"sync"
	"time"

	"appengine"
)

// Vertex of the graph
type Vertex struct {
	Type  string
	Value string
	Count int64
}

// Edge of the graph
type Edge struct {
	Nodes []string
}

// Miner takes posts and mines out a graph
func Miner(c appengine.Context, posts <-chan Data, workers int) (<-chan interface{}, <-chan interface{}) {
	vertexes := make(chan interface{}, 10000)
	edges := make(chan interface{}, 10000)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		go func(i int) {
			miner(c, posts, vertexes, edges, i)
			wg.Done()
		}(i)
	}
	wg.Add(workers)

	go func() {
		wg.Wait()
		close(vertexes)
		close(edges)
	}()
	return vertexes, edges
}

func miner(c appengine.Context, posts <-chan Data, vertexes chan<- interface{}, edges chan<- interface{}, i int) {
	var data struct {
		Tags []string `xml:"category"`
		Imgs []struct {
			URL string `xml:"url,attr"`
		} `xml:"content"`
	}

	for post := range posts {
		vertexes <- Vertex{"Pst", post.KEY, 0}

		// log.Printf("Miner %d: Got Post: %s", i, post.KEY)
		// log.Printf("Data: %s", post.XML)

		if err := xml.Unmarshal([]byte("<item>"+post.XML+"</item>"), &data); err != nil {
			c.Errorf("Miner %d: Error %s", i, err)
		}

		for _, tag := range data.Tags {
			// log.Printf("Found Tag: %s", tag)
			vertexes <- Vertex{"Tag", tag, 0}
			edges <- Edge{[]string{"Tag" + tag, "Pst" + post.KEY}}
		}

		for _, img := range data.Imgs {
			// log.Printf("Found Img: %s", img.URL)
			vertexes <- Vertex{"Img", img.URL, 0}
			edges <- Edge{[]string{"Img" + img.URL, "Pst" + post.KEY}}
		}
		time.Sleep(time.Second)
	}
}
