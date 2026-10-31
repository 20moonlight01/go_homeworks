package trip

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolationCode = "23505"

func (r *Repository) Create(ctx context.Context, trip domain.Trip) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := r.exec(ctx)

	sqlText, args, err := psqlBuilder.
		Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude",
			"end_latitude", "end_longitude",
			"price", "status", "started_at",
		).
		Values(
			trip.Id, trip.UserId, trip.DriverId,
			trip.StartPoint.Latitude, trip.StartPoint.Longitude,
			trip.EndPoint.Latitude, trip.EndPoint.Longitude,
			trip.Price, trip.Status, trip.StartedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if _, err := exec.Exec(ctx, sqlText, args...); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrDriverBusy
		}

		return fmt.Errorf("insert trip: %w", err)
	}

	historySqlText, hArgs, err := psqlBuilder.
		Insert("trip_status_history").
		Columns(
			"trip_id", "from_status", "to_status", "reason",
		).
		Values(
			trip.Id, nil, string(domain.StatusActive), "trip started",
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if _, err := exec.Exec(ctx, historySqlText, hArgs...); err != nil {
		return fmt.Errorf("insert trip status history: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgUniqueViolationCode
	}

	return false
}
