//go:build integration

package trip_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

func TestRepository_SavePosition_UpdatesLastPositionAt(t *testing.T) {
	pool := setupTestDB(t)
	r, txm := setupRepo(pool)
	ctx := context.Background()

	tr, err := domain.NewTrip(uuid.New(), uuid.New(), 55, 37, 56, 38, 100)
	require.NoError(t, err)
	tr.StartedAt = time.Now().UTC().Add(-time.Hour)

	insertTripRaw(t, pool, tr, nil)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM trips WHERE id = $1", tr.Id)
		require.NoError(t, err)
	})

	newTime := time.Now().UTC().Add(-time.Minute)
	oldTime := time.Now().UTC().Add(-30 * time.Minute)

	var saved *domain.Position
	err = txm.Do(ctx, func(ctx context.Context) error {
		var e error
		saved, e = r.SavePosition(ctx, &domain.Position{
			TripId: tr.Id, Latitude: 55, Longitude: 37, RecordedAt: newTime,
		})
		return e
	})
	require.NoError(t, err)
	require.NotNil(t, saved)

	err = txm.Do(ctx, func(ctx context.Context) error {
		_, e := r.SavePosition(ctx, &domain.Position{
			TripId: tr.Id, Latitude: 55, Longitude: 37, RecordedAt: oldTime,
		})
		return e
	})
	require.NoError(t, err)

	var lastPos *time.Time
	err = pool.QueryRow(ctx, "SELECT last_position_at FROM trips WHERE id = $1", tr.Id).Scan(&lastPos)
	require.NoError(t, err)
	require.NotNil(t, lastPos)
	assert.WithinDuration(t, newTime, *lastPos, time.Second, "last_position_at не должен уменьшиться")
}

func TestRepository_SavePosition_Rollback(t *testing.T) {
	pool := setupTestDB(t)
	r, txm := setupRepo(pool)
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION fail_position() RETURNS trigger AS $$
		BEGIN RAISE EXCEPTION 'forced failure for rollback test'; END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER fail_position_trigger BEFORE INSERT ON trip_positions
			FOR EACH ROW EXECUTE FUNCTION fail_position();
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `
			DROP TRIGGER IF EXISTS fail_position_trigger ON trip_positions;
			DROP FUNCTION IF EXISTS fail_position();
		`)
		require.NoError(t, err)
	})

	tr, err := domain.NewTrip(uuid.New(), uuid.New(), 55, 37, 56, 38, 100)
	require.NoError(t, err)
	insertTripRaw(t, pool, tr, nil)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM trips WHERE id = $1", tr.Id)
		require.NoError(t, err)
	})

	err = txm.Do(ctx, func(ctx context.Context) error {
		_, e := r.SavePosition(ctx, &domain.Position{
			TripId: tr.Id, Latitude: 55, Longitude: 37, RecordedAt: time.Now().UTC(),
		})
		return e
	})
	require.Error(t, err)

	var posCount int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM trip_positions WHERE trip_id = $1", tr.Id).Scan(&posCount)
	require.NoError(t, err)
	assert.Equal(t, 0, posCount, "trip_positions должна была откатиться")

	var lastPos *time.Time
	err = pool.QueryRow(ctx, "SELECT last_position_at FROM trips WHERE id = $1", tr.Id).Scan(&lastPos)
	require.NoError(t, err)
	assert.Nil(t, lastPos, "last_position_at не должен обновиться при откате")
}
