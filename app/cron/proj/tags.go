package proj

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/bign8/chive-show/app/cron/chain"

	"appengine"
	"appengine/memcache"
)

const tagsMemcacheKey = "tags-baby"

// Tags etrieves the tags from the dataset
func Tags(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		c.Infof("Time took: %v", time.Since(start))
	}()
	// w.Header().Set("Content-Type", "text/csv; charset=utf-8")

	// Check from memcache
	if item, err := memcache.Get(c, tagsMemcacheKey); err == nil {
		w.Write(item.Value)
		return
	}

	// Pretty sure this doesn't work on prod, but works awesome in dev
	runtime.GOMAXPROCS(runtime.NumCPU())
	tags := chain.FanOut(50, 10000, getItems(c), tags) // Pull and clean tags

	// Build a counter dictionary
	found := map[string]int64{}
	for tag := range tags {
		found[tag.(string)]++
	}

	// Compute average (used to clip data, so it's not huge)
	avg := int64(0)
	for _, value := range found {
		avg += value
	}
	avg /= int64(len(found))
	c.Infof("Num tags: %v; Avg: %v", len(found), avg)

	// Compute the 75%-tile
	cap := int64(0)
	for key, value := range found {
		if avg <= value {
			cap += value
		} else {
			delete(found, key)
		}
	}
	cap /= int64(len(found))
	c.Infof("Above average tags: %v; 75%%-tile: %v", len(found), cap)

	// Output results
	var buffer bytes.Buffer
	result := int64(0)
	for key, value := range found {
		if cap <= value {
			buffer.WriteString(fmt.Sprintf("%s,%d\n", key, value))
			result++
		}
	}
	data := buffer.Bytes()
	w.Write(data)
	c.Infof("Returned tags: %v", result)

	// Save to memcache, but only wait up to 3ms.
	done := make(chan bool, 1)
	go func() {
		memcache.Set(c, &memcache.Item{
			Key:   tagsMemcacheKey,
			Value: data,
		})
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Millisecond):
	}
}

func tags(obj interface{}, out chan<- interface{}, idx int) {
	for _, tag := range validTags((obj.(Item)).Tags) {
		out <- tag
	}
}

// http://xpo6.com/list-of-english-stop-words/
var chiveWords = "web only,thebrigade,theberry,thechive,chive,chive humanity,"
var stopWords = chiveWords + "a,able,about,across,after,all,almost,also,am,among,an,and,any,are,as,at,be,because,been,but,by,can,cannot,could,dear,did,do,does,either,else,ever,every,for,from,get,got,had,has,have,he,her,hers,him,his,how,however,i,if,in,into,is,it,its,just,least,let,like,likely,may,me,might,most,must,my,neither,no,nor,not,of,off,often,on,only,or,other,our,own,rather,said,say,says,she,should,since,so,some,than,that,the,their,them,then,there,these,they,this,tis,to,too,twas,us,wants,was,we,were,what,when,where,which,while,who,whom,why,will,with,would,yet,you,your"
var stops = map[string]bool{}

func validTags(tags []string) (res []string) {
	if len(stops) == 0 {
		for _, s := range strings.Split(stopWords, ",") {
			stops[s] = true
		}
	}
	for _, tag := range tags {
		tag = strings.ToLower(tag)
		if !stops[tag] {
			res = append(res, tag)
		}
	}
	return
}
