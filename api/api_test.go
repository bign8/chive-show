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

func TestHandleFail(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s := func(_ context.Context, opts *models.ListOptions) (*models.ListResult, error) {
		return nil, errors.New("fail")
	}
	handle(s).ServeHTTP(w, r)
}

func TestHandlePass(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Add("x-forwarded-host", "yeet")
	r.Header.Add("x-forwarded-proto", "yeet")
	w := httptest.NewRecorder()
	s := func(_ context.Context, opts *models.ListOptions) (*models.ListResult, error) {
		return &models.ListResult{
			Posts: make([]models.Post, opts.Count),
			Next: &models.ListOptions{
				Cursor: "curses",
				Tag:    "mixed",
			},
		}, nil
	}
	handle(s).ServeHTTP(w, r)
}
