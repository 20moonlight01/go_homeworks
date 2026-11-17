package trip

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

func (r *Repository) SaveFinished(ctx context.Context, t *domain.Trip, fromStatus domain.Status) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := r.exec(ctx)

	sqlText, args, err := psqlBuilder.
		Update("trips").
		Set("status", string(t.Status)).
		Set("finished_at", t.FinishedAt).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": t.Id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update query: %w", err)
	}

	tag, err := exec.Exec(ctx, sqlText, args...)
	if err != nil {
		return fmt.Errorf("update trip: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTripNotFound
	}

	historySqlText, hArgs, err := psqlBuilder.
		Insert("trip_status_history").
		Columns(
			"trip_id", "from_status", "to_status", "reason",
		).
		Values(
			t.Id, string(fromStatus), string(t.Status), "trip finished",
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
