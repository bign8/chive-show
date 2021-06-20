package appengine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"go.opencensus.io/trace"
)

// API design from https://pkg.go.dev/google.golang.org/appengine@v1.6.7/log
// Severity levels from https://pkg.go.dev/cloud.google.com/go/logging#pkg-constants

// Debug formats its arguments according to the format, analogous to fmt.Printf,
// and records the text as a log message at Debug level. The message will be associated
// with the request linked with the provided context.
//
// Debug means debug or trace information.
func Debug(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Debug, format, args...)
}

// Info means routine information, such as ongoing status or performance.
func Info(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Info, format, args...)
}

// Notice means normal but significant events, such as start up, shut down, or configuration.
func Notice(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Notice, format, args...)
}

// Warning means events that might cause problems.
func Warning(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Warning, format, args...)
}

// Error means events that are likely to cause problems.
func Error(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Error, format, args...)
}

// Critical means events that cause more severe problems or brief outages.
func Critical(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Critical, format, args...)
}

// Alert  means a person must take an action immediately.
func Alert(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Alert, format, args...)
}

// Emergency means one or more systems are unusable.
func Emergency(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, logging.Emergency, format, args...)
}

type trappedContextKey string

// Assigned to ctx in `server`
const (
	loggingKey = trappedContextKey("loggingKey")
	tracingKey = trappedContextKey("tracingKey")
)

func logf(ctx context.Context, level logging.Severity, format string, args ...interface{}) {
	payload := fmt.Sprintf(format, args...)
	logger, loggerOK := ctx.Value(loggingKey).(*logging.Logger)
	traceID, traceOK := ctx.Value(tracingKey).(string)
	if loggerOK && traceOK {
		// because unit tests don't have real contexts :cry:
		spanCtx := trace.FromContext(ctx).SpanContext()
		logger.Log(logging.Entry{
			Timestamp:    time.Now(),
			Severity:     level,
			Payload:      payload,
			Trace:        traceID,
			SpanID:       spanCtx.SpanID.String(),
			TraceSampled: spanCtx.IsSampled(),
		})
	}
	if !isSecondGen() {
		log.Printf(`[%s]: %s`, strings.ToUpper(level.String()[:4]), payload)
	}
}
