package api

import (
  "net/url"
  "testing"
)

func TestCountQueryNegative(t *testing.T) {
  test_url, err := url.Parse("http://example.com/sub/path/file.txt?as=df&count=-1&apples=oranges")
  if err != nil {
    t.Error("Some error %v", err)
  }
  test_value := get_url_count(test_url)
  if test_value != 2 {
    t.Error("Incorrect result 2 !=", test_value)
  }
}

func TestCountQueryPositive(t *testing.T) {
  test_url, err := url.Parse("http://example.com/sub/path/file.txt?as=df&count=99&apples=oranges")
  if err != nil {
    t.Error("Some error %v", err)
  }
  test_value := get_url_count(test_url)
  if test_value != 2 {
    t.Error("Incorrect result 2 !=", test_value)
  }
}
