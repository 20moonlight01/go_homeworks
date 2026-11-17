package finish_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/finish"
)

func TestUsecase_Execute(t *testing.T) {
	tripID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)
		logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
		mockMetrics := NewMockmetrics(ctrl)

		activeTrip := &trip.Trip{Id: tripID, Status: trip.StatusActive}

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().GetByIDForUpdate(gomock.Any(), tripID).Return(activeTrip, nil)
		mockRepo.EXPECT().SaveFinished(gomock.Any(), activeTrip, trip.StatusActive).Return(nil)
		mockMetrics.EXPECT().IncTripsCompleted(gomock.Any())

		uc := finish.NewUsecase(mockRepo, mockTx, logger, mockMetrics)
		result, err := uc.Execute(context.Background(), tripID)

		require.NoError(t, err)
		assert.Equal(t, tripID, result.Id)
		assert.Equal(t, trip.StatusCompleted, result.Status)
		assert.NotNil(t, result.FinishedAt)
	})

	t.Run("trip not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)
		logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
		mockMetrics := NewMockmetrics(ctrl)

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().GetByIDForUpdate(gomock.Any(), tripID).Return(nil, trip.ErrTripNotFound)

		uc := finish.NewUsecase(mockRepo, mockTx, logger, mockMetrics)
		_, err := uc.Execute(context.Background(), tripID)

		require.Error(t, err)
		assert.ErrorIs(t, err, trip.ErrTripNotFound)
	})

	t.Run("trip completed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)
		logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
		mockMetrics := NewMockmetrics(ctrl)

		completedTrip := &trip.Trip{Id: tripID, Status: trip.StatusCompleted}

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().GetByIDForUpdate(gomock.Any(), tripID).Return(completedTrip, nil)

		uc := finish.NewUsecase(mockRepo, mockTx, logger, mockMetrics)
		_, err := uc.Execute(context.Background(), tripID)

		require.Error(t, err)
		assert.ErrorIs(t, err, trip.ErrTripCompleted)
	})
}
