package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var target = flag.String("target", "", "<scheme>://<host>[:<port>] of serivce under test")

func fetch(tb testing.TB, addr string) io.ReadCloser {
	res, err := http.Get(addr)
	if err != nil {
		tb.Fatalf("Unable to make request: %v", err)
	}
	tb.Logf(" GET  %d %s", res.StatusCode, addr)
	if res.StatusCode != http.StatusOK {
		bits, _ := httputil.DumpResponse(res, true)
		tb.Log(string(bits))
		tb.Fatalf("Retrieved non-200 status: %d", res.StatusCode)
	}
	return res.Body
}

func fetchNextURL(tb testing.TB, addr string) string {
	body := fetch(tb, addr)
	defer body.Close()
	var miner struct {
		NextURL string        `json:"next_url"`
		Data    []interface{} `json:"data"`
	}
	if err := json.NewDecoder(body).Decode(&miner); err != nil {
		bits, _ := httputil.DumpResponse(&http.Response{Body: body}, true)
		tb.Log(string(bits))
		tb.Fatalf("Unable to process Payload: %v", err)
	}
	if len(miner.Data) != 3 {
		tb.Errorf("Expected 3 posts, retrieved %d", len(miner.Data))
	}
	return miner.NextURL
}

func TestRandomEndpoint(t *testing.T) {
	if *target == "" {
		t.Skip("missing `target` flag")
	}
	next := fetchNextURL(t, *target+"/api/v1/post/random?count=3")
	last := fetchNextURL(t, next)
	t.Logf("STOP %q", last)
}

func crawl(n *html.Node, queue chan string) {
	if n.Type == html.ElementNode && (n.Data == "a" || n.Data == "script" || n.Data == "link") {
		for _, a := range n.Attr {
			if a.Key == "href" || a.Key == "src" {
				queue <- a.Val
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		crawl(c, queue)
	}
}

func fetchLinks(tb testing.TB, addr string, queue chan string) {
	body := fetch(tb, addr)
	defer body.Close()
	doc, err := html.Parse(body)
	if err != nil {
		tb.Fatalf("Unable to parse body: %v", err)
	}
	crawl(doc, queue)
}

func TestForBrokenLinks(t *testing.T) {
	if *target == "" {
		t.Skip("missing `target` flag")
	}
	base, err := url.Parse(*target)
	if err != nil {
		t.Fatalf("Unable to process URL: %v", err)
	}
	queue := make(chan string, 64)
	queue <- *target
	hit := make(map[string]bool, 64)
	for {
		select {
		case item := <-queue:

			// Resolve relative URLs
			ref, err := url.Parse(item)
			if err != nil {
				t.Fatalf("Unable to process %s : %v", item, err)
			}
			item = base.ResolveReference(ref).String()

			// don't double check links
			if hit[item] {
				continue
			}
			hit[item] = true

			// If we are an internal link, mine for content
			if strings.HasPrefix(item, *target) {
				fetchLinks(t, item, queue)
				continue
			}

			// External links should not have redirects
			res, err := http.Head(item)
			if err != nil {
				t.Fatalf("Unable to load %s : %v", item, err)
			}
			t.Logf("HEAD %d %s", res.StatusCode, item)
			if res.StatusCode != http.StatusOK {
				t.Fatalf("non-200 external hop: %d", res.StatusCode)
			}
		default:
			t.Log("Link crawl complete!")
			return
		}
	}
}
