package logging

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// LoggerWrapper wraps slog to automatically extract trace IDs
type LoggerWrapper struct {
	logger *slog.Logger
}

// New initializes a structured JSON logger
func New() *LoggerWrapper {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &LoggerWrapper{logger: l}
}

// WithContext returns a logger enriched with trace_id and span_id if available
func (lw *LoggerWrapper) WithContext(ctx context.Context) *slog.Logger {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return lw.logger.With(
			"trace_id", span.SpanContext().TraceID().String(),
			"span_id", span.SpanContext().SpanID().String(),
		)
	}
	return lw.logger
}

// Get raw logger
func (lw *LoggerWrapper) Get() *slog.Logger {
	return lw.logger
}
