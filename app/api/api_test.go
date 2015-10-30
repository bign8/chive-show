package api

import (
	"net/url"
	"testing"
)

func TestCountQueryNegative(t *testing.T) {
	testURL, err := url.Parse("http://example.com/sub/path/file.txt?as=df&count=-1&apples=oranges")
	if err != nil {
		t.Errorf("Some error: %s", err)
	}
	testValue := getURLCount(testURL)
	if testValue != 2 {
		t.Error("Incorrect result 2 !=", testValue)
	}
}

func TestCountQueryPositive(t *testing.T) {
	testURL, err := url.Parse("http://example.com/sub/path/file.txt?as=df&count=99&apples=oranges")
	if err != nil {
		t.Errorf("Some error: %s", err)
	}
	testValue := getURLCount(testURL)
	if testValue != 2 {
		t.Error("Incorrect result 2 !=", testValue)
	}
}
