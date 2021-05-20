package api

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"cloud.google.com/go/datastore"
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

type fakeMultiGet func(context.Context, []*datastore.Key, interface{}) error

func (fmg fakeMultiGet) GetMulti(ctx context.Context, keys []*datastore.Key, data interface{}) error {
	return fmg(ctx, keys, data)
}

func TestRandom(t *testing.T) {
	var (
		err1 = errors.New("fail to fetch")
		err2 = errors.New("fail to multi")
	)
	m := fakeMultiGet(func(ctx context.Context, keys []*datastore.Key, dataRaw interface{}) error {
		data, ok := dataRaw.([]models.Post)
		if !ok {
			panic("not an array of models")
		}
		data[0].GUID = "trash"
		return err2
	})
	g := func(context.Context) ([]*datastore.Key, error) {
		return make([]*datastore.Key, 3), err1
	}
	r := httptest.NewRequest("GET", "/random", nil)

	// failed to fetch
	w := httptest.NewRecorder()
	random(m, g).ServeHTTP(w, r)

	// fail to multi
	err1 = nil
	w = httptest.NewRecorder()
	random(m, g).ServeHTTP(w, r)

	// okay!
	err2 = nil
	w = httptest.NewRecorder()
	random(m, g).ServeHTTP(w, r)

	// not enough
	r.URL.RawQuery = "count=4"
	// r.URL.Query().Add("count", "4")
	w = httptest.NewRecorder()
	random(m, g).ServeHTTP(w, r)
}
