package domain

import "errors"

var (
	ErrTripNotFound = errors.New("trip was not found")

	ErrTripCompleted = errors.New("trip has been already completed")

	ErrDriverBusy = errors.New("driver already has active trip")
)
