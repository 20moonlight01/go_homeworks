package trip

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/model"
)

func (r *Repository) SelectStale(ctx context.Context, staleAfter time.Duration) ([]model.StaleTrip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := r.exec(ctx)

	sqlText, args, err := psqlBuilder.
		Select("id", "driver_id", "started_at", "last_position_at").
		From("trips").
		Where(sq.Eq{"status": string(domain.StatusActive)}).
		Where(sq.Expr("COALESCE(last_position_at, started_at) < now() - ?::interval", staleAfter.String())).
		OrderBy("COALESCE(last_position_at, started_at) ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select stale query: %w", err)
	}

	rows, err := exec.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("select stale trips: %w", err)
	}
	defer rows.Close()

	var result []model.StaleTrip
	for rows.Next() {
		var s model.StaleTrip
		if err := rows.Scan(&s.Id, &s.DriverId, &s.StartedAt, &s.LastPositionAt); err != nil {
			return nil, fmt.Errorf("scan stale trip: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return result, nil
}
