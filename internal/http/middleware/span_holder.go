package middleware

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type spanHolderKey struct{}

type spanHolder struct {
	span trace.Span
}

func contextWithSpanHolder(ctx context.Context) (context.Context, *spanHolder) {
	h := &spanHolder{}
	return context.WithValue(ctx, spanHolderKey{}, h), h
}

func setSpanInHolder(ctx context.Context, span trace.Span) {
	if h, ok := ctx.Value(spanHolderKey{}).(*spanHolder); ok && h != nil {
		h.span = span
	}
}
