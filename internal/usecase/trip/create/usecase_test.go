package create_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/infra/telemetry"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/create"
)

func TestUsecase_Execute(t *testing.T) {
	validReq := create.Request{
		UserId:         uuid.New(),
		DriverId:       uuid.New(),
		StartLatitude:  55.0,
		StartLongitude: 37.0,
		EndLatitude:    56.0,
		EndLongitude:   38.0,
		Price:          100,
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)
		logger := telemetry.NewLogger(slog.LevelInfo)
		metrics := telemetry.NewNoopMetrics()

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

		uc := create.NewUsecase(mockRepo, mockTx, logger, metrics)
		tripData, err := uc.Execute(context.Background(), validReq)

		require.NoError(t, err)
		assert.NotNil(t, tripData)
		assert.Equal(t, trip.StatusActive, tripData.Status)
		assert.Equal(t, validReq.UserId, tripData.UserId)
		assert.Equal(t, validReq.DriverId, tripData.DriverId)
	})

	t.Run("validation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)
		logger := telemetry.NewLogger(slog.LevelInfo)
		metrics := telemetry.NewNoopMetrics()

		uc := create.NewUsecase(mockRepo, mockTx, logger, metrics)

		invalidReq := validReq
		invalidReq.Price = -1

		_, err := uc.Execute(context.Background(), invalidReq)

		require.Error(t, err)
		var valErr *trip.ValidationError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "price", valErr.Field)
	})

	t.Run("driver busy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)
		logger := telemetry.NewLogger(slog.LevelInfo)
		metrics := telemetry.NewNoopMetrics()

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(trip.ErrDriverBusy)

		uc := create.NewUsecase(mockRepo, mockTx, logger, metrics)
		_, err := uc.Execute(context.Background(), validReq)

		require.Error(t, err)
		assert.ErrorIs(t, err, trip.ErrDriverBusy)
	})
}
