package models

import (
	"bytes"
	"compress/gzip"
	"testing"
)

func chk(t testing.TB, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(`[]`))
	w.Close()
	x := &Post{
		MediaBytes: b.Bytes(),
	}
	chk(t, x.Load(nil))
}

func TestSave(t *testing.T) {
	x := &Post{
		Media: []Media{{
			Title: "hello",
		}},
	}
	_, err := x.Save()
	chk(t, err)
}
