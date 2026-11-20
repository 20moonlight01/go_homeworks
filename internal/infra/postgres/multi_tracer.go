package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type multiTracer struct {
	tracers []pgx.QueryTracer
}

func NewMultiTracer(tracers ...pgx.QueryTracer) pgx.QueryTracer {
	return &multiTracer{tracers: tracers}
}

func (m *multiTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, t := range m.tracers {
		ctx = t.TraceQueryStart(ctx, conn, data)
	}
	return ctx
}

func (m *multiTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for i := len(m.tracers) - 1; i >= 0; i-- {
		m.tracers[i].TraceQueryEnd(ctx, conn, data)
	}
}
