package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/20moonlight01/go_homeworks/internal/db"
	"github.com/20moonlight01/go_homeworks/internal/domain"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const pgUniqueViolationCode = "23505"

var psqlBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type TripRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TripRepository {
	return &TripRepository{
		pool:         pool,
		queryTimeout: queryTimeout,
	}
}

func (r *TripRepository) Create(ctx context.Context, trip api.Trip) (api.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.GetExecutor(ctx, r.pool)

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
		return api.Trip{}, fmt.Errorf("build insert query: %w", err)
	}

	if _, err := exec.Exec(ctx, sqlText, args...); err != nil {
		if isUniqueViolation(err) {
			return api.Trip{}, domain.ErrDriverBusy
		}

		return api.Trip{}, fmt.Errorf("insert trip: %w", err)
	}

	historySqlText, hArgs, err := psqlBuilder.
		Insert("trip_status_history").
		Columns(
			"trip_id", "from_status", "to_status", "reason",
		).
		Values(
			trip.Id, nil, string(api.Active), "trip started",
		).
		ToSql()
	if err != nil {
		return api.Trip{}, fmt.Errorf("build insert query: %w", err)
	}

	if _, err := exec.Exec(ctx, historySqlText, hArgs...); err != nil {
		return api.Trip{}, fmt.Errorf("insert trip status history: %w", err)
	}

	return trip, nil
}

func (r *TripRepository) GetByID(ctx context.Context, id openapi_types.UUID) (api.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.GetExecutor(ctx, r.pool)

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
		return api.Trip{}, fmt.Errorf("build select query: %w", err)
	}

	var (
		trip       api.Trip
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
			return api.Trip{}, domain.ErrTripNotFound
		}

		return api.Trip{}, fmt.Errorf("select trip: %w", err)
	}

	trip.Status = api.TripStatus(status)
	trip.FinishedAt = finishedAt

	return trip, nil
}

func (r *TripRepository) Finish(ctx context.Context, id openapi_types.UUID) (api.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.GetExecutor(ctx, r.pool)

	sqlText, args, err := psqlBuilder.
		Update("trips").
		Set("status", string(api.Completed)).
		Set("finished_at", sq.Expr("now()")).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id, "status": string(api.Active)}).
		ToSql()
	if err != nil {
		return api.Trip{}, fmt.Errorf("build update query: %w", err)
	}

	tag, err := exec.Exec(ctx, sqlText, args...)
	if err != nil {
		return api.Trip{}, fmt.Errorf("update trip: %w", err)
	}

	if tag.RowsAffected() == 0 {
		_, err = r.GetByID(ctx, id)

		if errors.Is(err, domain.ErrTripNotFound) {
			return api.Trip{}, domain.ErrTripNotFound
		}

		if err != nil {
			return api.Trip{}, err
		}

		return api.Trip{}, domain.ErrTripCompleted
	}

	historySqlText, hArgs, err := psqlBuilder.
		Insert("trip_status_history").
		Columns(
			"trip_id", "from_status", "to_status", "reason",
		).
		Values(
			id, string(api.Active), string(api.Completed), "trip finished",
		).
		ToSql()
	if err != nil {
		return api.Trip{}, fmt.Errorf("build insert query: %w", err)
	}

	if _, err := exec.Exec(ctx, historySqlText, hArgs...); err != nil {
		return api.Trip{}, fmt.Errorf("insert trip status history: %w", err)
	}

	return r.GetByID(ctx, id)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgUniqueViolationCode
	}

	return false
}
