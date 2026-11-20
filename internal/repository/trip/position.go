package trip

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

func (r *Repository) SavePosition(ctx context.Context, p *domain.Position) (*domain.Position, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := r.exec(ctx)

	sqlText, args, err := psqlBuilder.
		Insert("trip_positions").
		Columns("trip_id", "latitude", "longitude", "recorded_at").
		Values(p.TripId, p.Latitude, p.Longitude, p.RecordedAt).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert position query: %w", err)
	}

	var id int64
	var createdAt time.Time
	if err := exec.QueryRow(ctx, sqlText, args...).Scan(&id, &createdAt); err != nil {
		return nil, fmt.Errorf("insert position: %w", err)
	}

	updateSql, updateArgs, err := psqlBuilder.
		Update("trips").
		Set("last_position_at", sq.Expr(
			"GREATEST(COALESCE(last_position_at, ?), ?)", p.RecordedAt, p.RecordedAt,
		)).
		Where(sq.Eq{"id": p.TripId}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update last_position_at query: %w", err)
	}

	tag, err := exec.Exec(ctx, updateSql, updateArgs...)
	if err != nil {
		return nil, fmt.Errorf("update last_position_at: %w", err)
	}
	if tag.RowsAffected() == 0 {
		t, getErr := r.GetByID(ctx, p.TripId)
		if getErr != nil {
			return nil, getErr
		}
		if t.Status != domain.StatusActive {
			return nil, domain.ErrTripCompleted
		}
		return nil, fmt.Errorf("update last_position_at: no rows affected")
	}

	p.Id = id
	p.CreatedAt = createdAt
	return p, nil
}

func (r *Repository) ListPositions(ctx context.Context, tripId uuid.UUID) ([]*domain.Position, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := r.exec(ctx)

	sqlText, args, err := psqlBuilder.
		Select(
			"id", "trip_id", "latitude", "longitude", "recorded_at", "created_at",
		).
		From("trip_positions").
		Where(sq.Eq{"trip_id": tripId}).
		OrderBy("recorded_at ASC", "id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select positions query: %w", err)
	}

	rows, err := exec.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("select positions: %w", err)
	}
	defer rows.Close()

	var positions []*domain.Position
	for rows.Next() {
		var p domain.Position
		if err := rows.Scan(
			&p.Id,
			&p.TripId,
			&p.Latitude,
			&p.Longitude,
			&p.RecordedAt,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		positions = append(positions, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return positions, nil
}
