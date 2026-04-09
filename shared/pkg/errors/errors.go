// Package errors provides unified error handling for all microservices.
package errors

import (
	"errors"
	"fmt"
)

// Standard error types
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when authorization fails
	ErrForbidden = errors.New("forbidden")

	// ErrInternal is returned for internal server errors
	ErrInternal = errors.New("internal server error")

	// ErrConflict is returned when a conflict occurs (e.g., concurrent update)
	ErrConflict = errors.New("conflict")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")
)

// AppError represents an application error with additional context.
type AppError struct {
	// Code is the error code for programmatic handling
	Code string
	// Message is the human-readable error message
	Message string
	// Err is the underlying error
	Err error
	// Details contains additional error details
	Details map[string]interface{}
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError.
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetail adds a detail to the error.
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Common error constructors
func NotFound(message string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: message,
		Err:     ErrNotFound,
	}
}

func AlreadyExists(message string) *AppError {
	return &AppError{
		Code:    "ALREADY_EXISTS",
		Message: message,
		Err:     ErrAlreadyExists,
	}
}

func InvalidInput(message string) *AppError {
	return &AppError{
		Code:    "INVALID_INPUT",
		Message: message,
		Err:     ErrInvalidInput,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Err:     ErrUnauthorized,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Code:    "FORBIDDEN",
		Message: message,
		Err:     ErrForbidden,
	}
}

func Internal(message string) *AppError {
	return &AppError{
		Code:    "INTERNAL",
		Message: message,
		Err:     ErrInternal,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Err:     ErrConflict,
	}
}

func Timeout(message string) *AppError {
	return &AppError{
		Code:    "TIMEOUT",
		Message: message,
		Err:     ErrTimeout,
	}
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if an error is an already exists error.
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInput checks if an error is an invalid input error.
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is a forbidden error.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsConflict checks if an error is a conflict error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsTimeout checks if an error is a timeout error.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}
