package cron

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/bign8/chive-show/models"
)

type rt func(*http.Request) (*http.Response, error)

func (fn rt) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

const miningFixture = `
<figure data-attachment-id="123"><img src="lead-in"></figure>
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
	trans := client.Transport
	defer func() { client.Transport = trans }()
	client.Transport = rt(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(miningFixture)),
			Request:    r,
		}, nil
	})
	post := models.Post{Link: "testing"}
	if err := mine(context.TODO(), logDefault(), &post); err != nil {
		t.Fatal(err)
	}
	if len(post.Media) != 3 {
		t.Fatalf("Expected 3 media, got %d", len(post.Media))
	}
}

func ExampleMineHandler() {
	miner := MineHandler(nil)
	data := url.Values{}
	data.Set("id", "1234")
	r := httptest.NewRequest(`GET`, `/cron/mine`, strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	miner.ServeHTTP(w, r)
	// x := &feedParser{
	// 	context: r.Context(),
	// }
	// // _, posts, err := x.getAndParseFeed(1)
	// post := models.Post{
	// 	Link: "https://thechive.com/2021/06/06/all-they-had-to-do-was-change-a-websites-phone-number-to-avoid-this-revenge/",
	// }
	// err := x.mine(&post)
	// if err != nil {
	// 	panic(err)
	// }
	// enc := json.NewEncoder(w)
	// enc.SetIndent(``, ` `)
	// enc.SetEscapeHTML(false)
	// enc.Encode(post)
}

func TestMineFull(t *testing.T) {
	l := log.New(os.Stderr, ``, 0)
	post, err := mineFull(context.Background(), l, 3701780)
	if err != nil {
		t.Fatalf("Unable to mine full: %v", err)
	}
	t.Logf("TODO: make assertions on %v", post)
}
