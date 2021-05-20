package api

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bign8/chive-show/models"
)

func TestCountQuery(t *testing.T) {
	cases := []struct {
		Name string
		URL  string
		Out  int
	}{
		{"negative", "count=-1", 2},
		{"positive", "count=99", 2},
		{"normal", "count=10", 10},
	}
	for _, test := range cases {
		t.Run(test.Name, func(tt *testing.T) {
			testURL, _ := url.Parse("http://example.com/sub/path/file.txt?as=df&" + test.URL + "&apples=oranges")
			got := getURLCount(testURL)
			if got != test.Out {
				tt.Errorf("wanted %d got %d", test.Out, got)
			}
		})
	}
}

type testLister func(int) ([]models.Post, error)

func (list testLister) ListPosts(ctx context.Context, count int) ([]models.Post, error) {
	return list(count)
}

func TestRandomFail(t *testing.T) {
	r := httptest.NewRequest("GET", "/random", nil)
	w := httptest.NewRecorder()
	s := func(count int) ([]models.Post, error) {
		return nil, errors.New("fail")
	}
	random(testLister(s)).ServeHTTP(w, r)
}

func TestRandomPass(t *testing.T) {
	r := httptest.NewRequest("GET", "/random", nil)
	w := httptest.NewRecorder()
	s := func(count int) ([]models.Post, error) {
		return make([]models.Post, count), nil
	}
	random(testLister(s)).ServeHTTP(w, r)
}
