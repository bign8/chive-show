package cron

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/httpreplay"
	"google.golang.org/api/option"
)

const replayFilename = `cron.replay`

var record = flag.Bool(`record`, false, `Record test data`)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	old := client
	defer func() { client = old }() // reset

	flag.Parse()
	if testing.Short() || !*record { // disabling external calls for now (drop the 2nd case to match datastore tests here https://github.com/googleapis/google-cloud-go/blob/datastore/v1.5.0/datastore/integration_test.go#L71)
		if *record {
			log.Fatal("Cannot combine -short and -record")
		}
		rep, err := httpreplay.NewReplayer(replayFilename)
		if err != nil {
			log.Fatal(err)
		}
		c, err := rep.Client(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		client = c
	} else if *record {
		rec, err := httpreplay.NewRecorder(replayFilename, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := rec.Close(); err != nil {
				log.Fatalf("closing recorder: %v", err)
			}
		}()
		c, err := rec.Client(context.Background(), option.WithoutAuthentication())
		if err != nil {
			log.Fatal(err)
		}
		client = c
	}

	return m.Run()
}
