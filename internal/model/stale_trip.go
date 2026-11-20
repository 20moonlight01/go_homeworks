package model

import (
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type StaleTrip struct {
	Id             uuid.UUID
	DriverId       uuid.UUID
	StartedAt      time.Time
	LastPositionAt *time.Time

	CheckSpanContext trace.SpanContext
}
