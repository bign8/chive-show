package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/bign8/chive-show/models"
)

var target = flag.String("target", "", "<scheme>://<host>[:<port>] of serivce under test")

func fetch(tb testing.TB, addr string) string {
	tb.Logf("GET %q", addr)
	res, err := http.Get(addr)
	if err != nil {
		tb.Fatalf("Unable to make request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		bits, _ := httputil.DumpResponse(res, true)
		tb.Log(string(bits))
		tb.Fatalf("Retrieved non-200 status: %d", res.StatusCode)
	}
	defer res.Body.Close()
	var miner struct {
		NextURL string        `json:"next_url"`
		Data    []models.Post `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&miner); err != nil {
		bits, _ := httputil.DumpResponse(res, true)
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
		t.Skip("missing `target` parameter")
	}
	next := fetch(t, *target+"/api/v1/post/random?count=3")
	last := fetch(t, next)
	t.Logf("STOP %q", last)
}
