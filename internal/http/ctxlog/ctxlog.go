package ctxlog

import "context"

type key struct{}

type Entry struct{ Err error }

func WithEntry(ctx context.Context, e *Entry) context.Context {
	return context.WithValue(ctx, key{}, e)
}

func FromContext(ctx context.Context) *Entry {
	e, _ := ctx.Value(key{}).(*Entry)
	return e
}
