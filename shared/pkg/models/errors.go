package models

import "errors"

// Model-specific errors.
var (
	// ErrInvalidCheckType is returned when an invalid check type is provided.
	ErrInvalidCheckType = errors.New("invalid check type")

	// ErrInvalidCheckStatus is returned when an invalid check status is provided.
	ErrInvalidCheckStatus = errors.New("invalid check status")

	// ErrInvalidEventStatus is returned when an invalid event status is provided.
	ErrInvalidEventStatus = errors.New("invalid event status")

	// ErrInvalidTaskStatus is returned when an invalid task status is provided.
	ErrInvalidTaskStatus = errors.New("invalid task status")

	// ErrEventNotFound is returned when an event is not found.
	ErrEventNotFound = errors.New("event not found")

	// ErrTaskNotFound is returned when a task is not found.
	ErrTaskNotFound = errors.New("task not found")

	// ErrInvalidResourceType is returned when an invalid resource type is provided.
	ErrInvalidResourceType = errors.New("invalid resource type")
)
