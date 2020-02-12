package models

import "testing"

func chk(t testing.TB, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	x := &Post{
		MediaBytes: []byte(`[]`),
	}
	chk(t, x.Load(nil))
}

func TestSave(t *testing.T) {
	x := &Post{
		Media: []Img{{
			Title: "hello",
		}},
	}
	_, err := x.Save()
	chk(t, err)
}
