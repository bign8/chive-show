package appengine

import (
	"context"
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func AppID(context.Context) string {
	return "crucial-alpha-706"
}

func Main() {

	// Register some views (TODO: figure out how these get reported)
	if err := view.Register(
		ochttp.ClientSentBytesDistribution,
		ochttp.ClientReceivedBytesDistribution,
		ochttp.ClientRoundtripLatencyDistribution,
		ochttp.ClientCompletedCount,
		ochttp.ServerRequestCountView,
		ochttp.ServerRequestBytesView,
		ochttp.ServerResponseBytesView,
		ochttp.ServerLatencyView,
		ochttp.ServerRequestCountByMethod,
		ochttp.ServerResponseCountByStatusCode,
	); err != nil {
		log.Fatal(err)
	}

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: AppID(context.TODO()),
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

	// Create parent trace spans for all incoming requests (include request metadata)
	format := propagation.HTTPFormat{}
	tracer := func(w http.ResponseWriter, r *http.Request) {
		name := r.Method + " " + r.URL.Path
		var (
			span *trace.Span
			ctx  context.Context
		)
		if parent, ok := format.SpanContextFromRequest(r); ok {
			ctx, span = trace.StartSpanWithRemoteParent(r.Context(), name, parent)
		} else {
			ctx, span = trace.StartSpan(r.Context(), name)
		}
		defer span.End()
		attrs := []trace.Attribute{
			trace.StringAttribute("URL", r.URL.String()),
		}
		for key := range r.Header {
			attrs = append(attrs, trace.StringAttribute(key, r.Header.Get(key)))
		}
		span.AddAttributes(attrs...)
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
