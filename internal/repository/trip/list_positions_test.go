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

func TestRepository_ListPositions_Sorted(t *testing.T) {
	pool := setupTestDB(t)
	r, _ := setupRepo(pool)
	ctx := context.Background()

	tr, err := domain.NewTrip(uuid.New(), uuid.New(), 55, 37, 56, 38, 100)
	require.NoError(t, err)
	insertTripRaw(t, pool, tr, nil)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM trips WHERE id = $1", tr.Id)
		require.NoError(t, err)
	})

	t1 := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	t2 := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Microsecond)
	t3 := time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Microsecond)

	for _, rt := range []time.Time{t1, t2, t3} {
		_, err := pool.Exec(ctx, `
			INSERT INTO trip_positions (trip_id, latitude, longitude, recorded_at)
			VALUES ($1, $2, $3, $4)
		`, tr.Id, 55.0, 37.0, rt)
		require.NoError(t, err)
	}

	positions, err := r.ListPositions(ctx, tr.Id)
	require.NoError(t, err)
	require.Len(t, positions, 3)

	assert.True(t, positions[0].RecordedAt.Before(positions[1].RecordedAt))
	assert.True(t, positions[1].RecordedAt.Before(positions[2].RecordedAt))
	assert.Equal(t, t3.UnixMicro(), positions[0].RecordedAt.UnixMicro())
	assert.Equal(t, t1.UnixMicro(), positions[1].RecordedAt.UnixMicro())
	assert.Equal(t, t2.UnixMicro(), positions[2].RecordedAt.UnixMicro())
}
