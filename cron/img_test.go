package cron

import (
	"image"
	"net/http"
	"testing"
	"time"

	// _ "image/gif"
	_ "image/jpeg"
	// _ "image/png"
)

func clear(tb testing.TB, err error) {
	if err != nil {
		tb.Fatal(err)
	}
}

func TestImagez(t *testing.T) {
	a := time.Now()
	// img := "https://thechive.com/wp-content/uploads/2021/06/4JZDQS4.jpg" // 600x627 in 256 bytes
	// img := "https://thechive.com/wp-content/uploads/2021/06/lead1-13.jpg" // 1200x628 in 512 bytes
	// img := "https://thechive.com/wp-content/uploads/2021/06/0ip0ybn7pgy61.jpg" // 600x872 in 256 bytes
	// img := "https://thechive.com/wp-content/uploads/2021/06/0t6VhTI.jpg" // 600x1066 in 256 bytes
	img := "https://thechive.com/wp-content/uploads/2021/06/7rtws32el3471.jpg" // 600x587 256 bytes
	req, err := http.NewRequest(http.MethodGet, img, nil)
	clear(t, err)
	req.Header.Add("Range", "bytes=0-1023") // seems like we can parse w/minimal data
	res, err := http.DefaultClient.Do(req)
	clear(t, err)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
		t.Fatalf("Bad response: %d", res.StatusCode)
	}
	cfg, _, err := image.DecodeConfig(res.Body)
	clear(t, err)
	t.Logf("Got Metrics: %#v", cfg)
	t.Logf("Read %d bytes in %s", -1, time.Since(a))
	clear(t, res.Body.Close())
}
