package logging

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

type LoggerWrapper struct {
	logger *slog.Logger
}

func New() *LoggerWrapper {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &LoggerWrapper{logger: l}
}

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

func (lw *LoggerWrapper) Get() *slog.Logger {
	return lw.logger
}
