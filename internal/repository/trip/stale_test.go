//go:build integration

package trip_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

func TestRepository_SelectStale(t *testing.T) {
	pool := setupTestDB(t)
	r, _ := setupRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC()

	staleNoPos, err := domain.NewTrip(uuid.New(), uuid.New(), 55, 37, 56, 38, 100)
	require.NoError(t, err)
	staleNoPos.StartedAt = now.Add(-2 * time.Hour)

	freshWithPos, err := domain.NewTrip(uuid.New(), uuid.New(), 55, 37, 56, 38, 100)
	require.NoError(t, err)
	freshWithPos.StartedAt = now.Add(-2 * time.Hour)

	completed, err := domain.NewTrip(uuid.New(), uuid.New(), 55, 37, 56, 38, 100)
	require.NoError(t, err)
	completed.StartedAt = now.Add(-2 * time.Hour)
	require.NoError(t, completed.Finish(now.Add(-time.Hour)))

	insertTripRaw(t, pool, staleNoPos, nil)
	insertTripRaw(t, pool, freshWithPos, &now)
	insertTripRaw(t, pool, completed, nil)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(),
			"DELETE FROM trips WHERE id IN ($1, $2, $3)",
			staleNoPos.Id, freshWithPos.Id, completed.Id)
		require.NoError(t, err)
	})

	result, err := r.SelectStale(ctx, 60*time.Second)
	require.NoError(t, err)

	ids := make([]uuid.UUID, 0, len(result))
	for _, s := range result {
		ids = append(ids, s.Id)
	}

	assert.Contains(t, ids, staleNoPos.Id, "stale active без координат должна попасть")
	assert.NotContains(t, ids, freshWithPos.Id, "свежая active не должна попасть")
	assert.NotContains(t, ids, completed.Id, "completed не должна попасть")
}

func insertTripRaw(t *testing.T, pool *pgxpool.Pool, tr *domain.Trip, lastPositionAt *time.Time) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		INSERT INTO trips (
			id, user_id, driver_id,
			start_latitude, start_longitude,
			end_latitude, end_longitude,
			price, status, started_at, finished_at, last_position_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		tr.Id, tr.UserId, tr.DriverId,
		tr.StartPoint.Latitude, tr.StartPoint.Longitude,
		tr.EndPoint.Latitude, tr.EndPoint.Longitude,
		tr.Price, string(tr.Status), tr.StartedAt, tr.FinishedAt, lastPositionAt,
	)
	require.NoError(t, err)
}
