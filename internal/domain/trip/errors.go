package trip

import "errors"

var (
	ErrTripNotFound  = errors.New("trip was not found")
	ErrTripCompleted = errors.New("trip has been already completed")
	ErrDriverBusy    = errors.New("driver already has active trip")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

const msgOutOfRange = "out of range"
