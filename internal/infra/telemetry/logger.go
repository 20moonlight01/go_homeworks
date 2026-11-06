package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

type traceHandler struct {
	slog.Handler
}

func NewLogger(level slog.Level) *slog.Logger {
	fmt.Print("DEBUG")
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	// return slog.New(h).With("service", "trip-service")
	return slog.New(&traceHandler{h}).With("service", "trip-service")
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)

	if span.SpanContext().IsValid() {
		r = r.Clone()
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	return h.Handler.Handle(ctx, r)
}
