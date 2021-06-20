package appengine

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
)

func ProjectID() string {
	// Note: there is also a metadata API service where other info is available
	// https://pkg.go.dev/cloud.google.com/go@v0.84.0/compute/metadata#ProjectID
	return os.Getenv(`GOOGLE_CLOUD_PROJECT`)
}

// https://github.com/golang/appengine/blob/856ef3e566899d9d74140c595bdf4791a1cbdc46/internal/identity.go#L40
func isSecondGen() bool {
	// Second-gen runtimes set $GAE_ENV so we use that to check if we're on a second-gen runtime.
	return os.Getenv("GAE_ENV") == "standard"
}

func Main() {

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: ProjectID(),
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

	// Start the http server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, server(ProjectID())); err != nil {
		log.Fatal(err)
	}
}

func server(parent string) http.HandlerFunc {

	// initialized logging client once
	logClient, err := logging.NewClient(context.Background(), parent)
	if err != nil {
		log.Fatal(err)
	}
	format := propagation.HTTPFormat{}

	return func(w http.ResponseWriter, r *http.Request) {

		// Create parent trace spans for all incoming requests (include request metadata)
		name := r.Method + " " + r.URL.Path
		var (
			span *trace.Span
			ctx  context.Context
		)
		if parent, ok := format.SpanContextFromRequest(r); ok {
			ctx, span = trace.StartSpanWithRemoteParent(r.Context(), name, parent)
		} else {
			// TODO: when serving on codespaces, try and use X-Request-ID for trace ID if possible
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
		spanCtx := span.SpanContext()

		// Create the appropriate traceID to link LOGS with TRACES
		// https://cloud.google.com/trace/docs/trace-log-integration#associating
		traceID := "projects/" + parent + "/traces/" + spanCtx.TraceID.String()
		ctx = context.WithValue(ctx, tracingKey, traceID)

		// Logging initialization
		logger := logClient.Logger(`appengine`, logging.ContextFunc(func() (context.Context, func()) {
			return context.WithTimeout(ctx, time.Second) // attempt to flush logs for 1 second max
		}))
		defer logger.Flush()
		ctx = context.WithValue(ctx, loggingKey, logger)

		// Forward the context outward (todo: remember response status on parent log)
		http.DefaultServeMux.ServeHTTP(w, r.WithContext(ctx))

		// Log to denote the overall request / response
		logger.Log(logging.Entry{
			Trace:        traceID,
			SpanID:       spanCtx.SpanID.String(),
			TraceSampled: spanCtx.IsSampled(),
			// TODO: set the severity level based on worst log during request
			HTTPRequest: &logging.HTTPRequest{
				Request: r,
				Status:  http.StatusOK,
				// TODO: more propreties can be set here
			},
		})
	}
}
