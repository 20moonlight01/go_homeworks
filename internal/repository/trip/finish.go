package trip

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

func (r *Repository) Finish(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := r.exec(ctx)

	sqlText, args, err := psqlBuilder.
		Update("trips").
		Set("status", string(domain.StatusCompleted)).
		Set("finished_at", sq.Expr("now()")).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id, "status": string(domain.StatusActive)}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update query: %w", err)
	}

	tag, err := exec.Exec(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("update trip: %w", err)
	}

	if tag.RowsAffected() == 0 {
		_, err = r.GetByID(ctx, id)

		if errors.Is(err, domain.ErrTripNotFound) {
			return nil, domain.ErrTripNotFound
		}

		if err != nil {
			return nil, err
		}

		return nil, domain.ErrTripCompleted
	}

	historySqlText, hArgs, err := psqlBuilder.
		Insert("trip_status_history").
		Columns(
			"trip_id", "from_status", "to_status", "reason",
		).
		Values(
			id, string(domain.StatusActive), string(domain.StatusCompleted), "trip finished",
		).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert query: %w", err)
	}

	if _, err := exec.Exec(ctx, historySqlText, hArgs...); err != nil {
		return nil, fmt.Errorf("insert trip status history: %w", err)
	}

	return r.GetByID(ctx, id)
}
