//go:build integration

package trip_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/infra/postgres/transaction"
	repo "github.com/20moonlight01/go_homeworks/internal/repository/trip"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	return pool
}

func setupRepo(pool *pgxpool.Pool) (*repo.Repository, transaction.TxManager) {
	txm := transaction.NewManager(pool)
	r := repo.NewRepository(pool, 5*time.Second)
	return r, txm
}

func TestRepository_SaveAndGet(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	r, txm := setupRepo(pool)
	ctx := context.Background()

	tr, err := trip.NewTrip(uuid.New(), uuid.New(), 55.0, 37.0, 56.0, 38.0, 100)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM trips WHERE id = $1", tr.Id)
	})

	err = txm.Do(ctx, func(ctx context.Context) error {
		return r.Save(ctx, tr)
	})
	require.NoError(t, err)

	fetched, err := r.GetByID(ctx, tr.Id)
	require.NoError(t, err)

	assert.Equal(t, tr.Id, fetched.Id)
	assert.Equal(t, tr.Status, fetched.Status)
}

func TestRepository_Save_TransactionRollback(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	r, txm := setupRepo(pool)
	ctx := context.Background()

	tr, _ := trip.NewTrip(uuid.New(), uuid.New(), 55.0, 37.0, 56.0, 38.0, 100)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM trips WHERE id = $1", tr.Id)
	})

	err := txm.Do(ctx, func(ctx context.Context) error {
		if err := r.Save(ctx, tr); err != nil {
			return err
		}
		return errors.New("искусственная ошибка для проверки отката")
	})
	require.Error(t, err)

	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM trips WHERE id = $1", tr.Id).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Транзакция должна была откатиться")
}

func TestRepository_Save_DriverBusy(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	r, txm := setupRepo(pool)
	ctx := context.Background()

	driverID := uuid.New()

	tr1, _ := trip.NewTrip(uuid.New(), driverID, 55.0, 37.0, 56.0, 38.0, 100)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM trips WHERE id IN ($1, $2)", tr1.Id, tr1.Id)
	})

	err := txm.Do(ctx, func(ctx context.Context) error {
		return r.Save(ctx, tr1)
	})
	require.NoError(t, err)

	tr2, _ := trip.NewTrip(uuid.New(), driverID, 55.0, 37.0, 56.0, 38.0, 100)
	err = txm.Do(ctx, func(ctx context.Context) error {
		return r.Save(ctx, tr2)
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, trip.ErrDriverBusy)
}
