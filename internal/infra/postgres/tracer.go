package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SpanTracer struct {
	inner pgx.QueryTracer
}

func NewSpanTracer(inner pgx.QueryTracer) *SpanTracer {
	return &SpanTracer{inner: inner}
}

func (t *SpanTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = t.inner.TraceQueryStart(ctx, conn, data)

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		op, table := parseSQL(data.SQL)
		if op != "" {
			span.SetAttributes(attribute.String("db.operation.name", op))
		}
		if table != "" {
			span.SetAttributes(attribute.String("db.collection.name", table))
		}
	}
	return ctx
}

func (t *SpanTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			span.SetAttributes(attribute.String("error.type", getErrorType(data.Err)))
			span.SetStatus(codes.Error, data.Err.Error())
		}
	}
	t.inner.TraceQueryEnd(ctx, conn, data)
}

func getErrorType(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return fmt.Sprintf("%T", err)
}

func parseSQL(sql string) (operation, table string) {
	sql = strings.TrimSpace(sql)
	upper := strings.ToUpper(sql)
	if strings.HasPrefix(upper, "INSERT INTO") {
		parts := strings.Fields(sql)
		if len(parts) >= 3 {
			return "INSERT", parts[2]
		}
		return "INSERT", ""
	}
	if strings.HasPrefix(upper, "UPDATE") {
		parts := strings.Fields(sql)
		if len(parts) >= 2 {
			return "UPDATE", parts[1]
		}
		return "UPDATE", ""
	}
	if strings.HasPrefix(upper, "SELECT") {
		idx := strings.Index(upper, "FROM ")
		if idx != -1 {
			rest := sql[idx+5:]
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				return "SELECT", parts[0]
			}
		}
		return "SELECT", ""
	}
	return "", ""
}
