package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"

	"github.com/bign8/chive-show/api"
	"github.com/bign8/chive-show/cron"
)

func main() {
	// http.HandleFunc("/", http.NotFound) // Default Handler
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/_ah/warmup", func(w http.ResponseWriter, r *http.Request) {
		// Needed to be able to migrate traffic on promotion.
		log.Println("Warmup Done")
		http.Error(w, "warm!", http.StatusOK)
	})

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: "crucial-alpha-706",
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	// By default, traces will be sampled relatively rarely. To change the
	// sampling frequency for your entire program, call ApplyConfig. Use a
	// ProbabilitySampler to sample a subset of traces, or use AlwaysSample to
	// collect a trace on every run.
	//
	// Be careful about using trace.AlwaysSample in a production application
	// with significant traffic: a new trace will be started and exported for
	// every request.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// https://cloud.google.com/docs/authentication/production
	// GOOGLE_APPLICATION_CREDENTIALS=<path-to>/service-account.json
	store, err := datastore.NewClient(context.Background(), "crucial-alpha-706")
	if err != nil {
		panic(err)
	}

	// Setup Other routes routes
	api.Init(store)
	cron.Init(store)

	// Create parent trace spans for all incoming requests (include request metadata)
	tracer := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := trace.StartSpan(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		http.DefaultServeMux.ServeHTTP(w, r.WithContext(ctx))
	}

	// Start the http server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, http.HandlerFunc(tracer)); err != nil {
		log.Fatal(err)
	}
}
