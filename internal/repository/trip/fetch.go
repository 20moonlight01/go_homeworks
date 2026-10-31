package trip

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := r.exec(ctx)

	sqlText, args, err := psqlBuilder.
		Select(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude",
			"end_latitude", "end_longitude",
			"price", "status",
			"started_at", "finished_at",
		).
		From("trips").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select query: %w", err)
	}

	var (
		trip       domain.Trip
		status     string
		finishedAt *time.Time
	)
	err = exec.
		QueryRow(ctx, sqlText, args...).
		Scan(
			&trip.Id,
			&trip.UserId,
			&trip.DriverId,
			&trip.StartPoint.Latitude,
			&trip.StartPoint.Longitude,
			&trip.EndPoint.Latitude,
			&trip.EndPoint.Longitude,
			&trip.Price,
			&status,
			&trip.StartedAt,
			&finishedAt,
		)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTripNotFound
		}

		return nil, fmt.Errorf("select trip: %w", err)
	}

	trip.Status = domain.Status(status)
	trip.FinishedAt = finishedAt

	return &trip, nil
}
