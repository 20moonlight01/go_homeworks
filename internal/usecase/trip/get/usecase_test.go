package get_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/get"
)

func TestUsecase_Execute(t *testing.T) {
	tripID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)

		expectedTrip := &trip.Trip{Id: tripID, Status: trip.StatusCompleted}

		mockRepo.EXPECT().GetByID(gomock.Any(), tripID).Return(expectedTrip, nil)

		uc := get.NewUsecase(mockRepo)
		result, err := uc.Execute(context.Background(), tripID)

		require.NoError(t, err)
		assert.Equal(t, expectedTrip, result)
	})

	t.Run("trip not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)

		mockRepo.EXPECT().GetByID(gomock.Any(), tripID).Return(nil, trip.ErrTripNotFound)

		uc := get.NewUsecase(mockRepo)
		_, err := uc.Execute(context.Background(), tripID)

		require.Error(t, err)
		assert.ErrorIs(t, err, trip.ErrTripNotFound)
	})

	t.Run("trip completed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMocktripRepository(ctrl)

		mockRepo.EXPECT().GetByID(gomock.Any(), tripID).Return(nil, trip.ErrTripCompleted)

		uc := get.NewUsecase(mockRepo)
		_, err := uc.Execute(context.Background(), tripID)

		require.Error(t, err)
		assert.ErrorIs(t, err, trip.ErrTripCompleted)
	})
}
