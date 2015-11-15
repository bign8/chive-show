package proj

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"appengine"
)

// Tags etrieves the tags from the dataset
func Tags(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	pages := puller(c)
	dirtyTags := getTags(c, pages, 100)
	tags := cleaner(dirtyTags)

	found := map[string]int64{}
	for tag := range tags {
		found[tag]++
	}

	for key, value := range found {
		fmt.Fprintf(w, "%s,%d\n", key, value)
	}

	c.Infof("Time took: %v", time.Since(start))
}

func getTags(c appengine.Context, in <-chan []byte, workers int) <-chan string {
	out := make(chan string, 10000)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(idx int) {
			tags(c, in, out, idx)
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func tags(c appengine.Context, in <-chan []byte, out chan<- string, idx int) {
	var xmlPage = XMLPage{}

	for data := range in {
		if err := xml.Unmarshal(data, &xmlPage); err != nil {
			c.Errorf("Miner %d: Error %s", idx, err)
			continue
		}

		for _, item := range xmlPage.Items {
			for _, tag := range item.Tags {
				out <- tag
			}
		}
	}
}

func cleaner(in <-chan string) <-chan string {
	// http://xpo6.com/list-of-english-stop-words/
	var stopWords = "a,able,about,across,after,all,almost,also,am,among,an,and,any,are,as,at,be,because,been,but,by,can,cannot,could,dear,did,do,does,either,else,ever,every,for,from,get,got,had,has,have,he,her,hers,him,his,how,however,i,if,in,into,is,it,its,just,least,let,like,likely,may,me,might,most,must,my,neither,no,nor,not,of,off,often,on,only,or,other,our,own,rather,said,say,says,she,should,since,so,some,than,that,the,their,them,then,there,these,they,this,tis,to,too,twas,us,wants,was,we,were,what,when,where,which,while,who,whom,why,will,with,would,yet,you,your"
	var stops = map[string]bool{}
	for _, s := range strings.Split(stopWords, ",") {
		stops[s] = true
	}

	out := make(chan string, 10000)
	go func() {
		for s := range in {
			s = strings.ToLower(s)
			if !stops[s] {
				out <- s
			}
		}
		close(out)
	}()
	return out
}
