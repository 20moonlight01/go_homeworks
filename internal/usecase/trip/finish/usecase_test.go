package finish_test

import (
	"context"
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

		expectedTrip := &trip.Trip{Id: tripID, Status: trip.StatusCompleted}

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().Finish(gomock.Any(), tripID).Return(expectedTrip, nil)

		uc := finish.NewUsecase(mockRepo, mockTx)
		result, err := uc.Execute(context.Background(), tripID)

		require.NoError(t, err)
		assert.Equal(t, expectedTrip, result)
	})

	t.Run("trip not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().Finish(gomock.Any(), tripID).Return(nil, trip.ErrTripNotFound)

		uc := finish.NewUsecase(mockRepo, mockTx)
		_, err := uc.Execute(context.Background(), tripID)

		require.Error(t, err)
		assert.ErrorIs(t, err, trip.ErrTripNotFound)
	})

	t.Run("trip completed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)
		mockTx := NewMocktxManager(ctrl)

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().Finish(gomock.Any(), tripID).Return(nil, trip.ErrTripCompleted)

		uc := finish.NewUsecase(mockRepo, mockTx)
		_, err := uc.Execute(context.Background(), tripID)

		require.Error(t, err)
		assert.ErrorIs(t, err, trip.ErrTripCompleted)
	})
}
