package cron

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bign8/chive-show/models"
)

type rt func(*http.Request) (*http.Response, error)

func (fn rt) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

const miningFixture = `
<figure><img src="lead-in"></figure>
<script id="chive-theme-js-js-extra">
var Some white space trash = {
	"items": [
		{"html": "<img src=\"js-embedded-1\">"},
		{"html": "<img src=\"js-embedded-2\">"}
	]
};
</script>
`

func TestMine(t *testing.T) {
	client.Transport = rt(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(miningFixture)),
			Request:    r,
		}, nil
	})
	post := models.Post{Link: "testing"}
	parser := feedParser{context: context.TODO()}
	if err := parser.mine(&post); err != nil {
		t.Fatal(err)
	}
	if len(post.Media) != 3 {
		t.Fatalf("Expected 3 media, got %d", len(post.Media))
	}
}

func TestDebug(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/cron/debug", nil)
	w := httptest.NewRecorder()
	debug(w, r)
}
