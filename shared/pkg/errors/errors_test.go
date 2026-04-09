package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/quality-gateway/shared/pkg/errors"
)

func TestAppError(t *testing.T) {
	t.Run("Error returns formatted message", func(t *testing.T) {
		err := errors.New("TEST_CODE", "test message")
		if err.Error() != "TEST_CODE: test message" {
			t.Errorf("expected 'TEST_CODE: test message', got '%s'", err.Error())
		}
	})

	t.Run("Error with underlying error", func(t *testing.T) {
		underlying := stderrors.New("underlying error")
		err := errors.Wrap(underlying, "TEST_CODE", "test message")
		expected := "TEST_CODE: test message: underlying error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := stderrors.New("underlying error")
		err := errors.Wrap(underlying, "TEST_CODE", "test message")
		if err.Unwrap() != underlying {
			t.Errorf("expected underlying error, got %v", err.Unwrap())
		}
	})

	t.Run("WithDetail adds details", func(t *testing.T) {
		err := errors.New("TEST_CODE", "test message")
		err = err.WithDetail("key", "value")
		if err.Details["key"] != "value" {
			t.Errorf("expected detail 'key' to be 'value', got %v", err.Details["key"])
		}
	})
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func(string) *errors.AppError
		expectedCode string
		checkFunc    func(error) bool
	}{
		{"NotFound", errors.NotFound, "NOT_FOUND", errors.IsNotFound},
		{"AlreadyExists", errors.AlreadyExists, "ALREADY_EXISTS", errors.IsAlreadyExists},
		{"InvalidInput", errors.InvalidInput, "INVALID_INPUT", errors.IsInvalidInput},
		{"Unauthorized", errors.Unauthorized, "UNAUTHORIZED", errors.IsUnauthorized},
		{"Forbidden", errors.Forbidden, "FORBIDDEN", errors.IsForbidden},
		{"Conflict", errors.Conflict, "CONFLICT", errors.IsConflict},
		{"Timeout", errors.Timeout, "TIMEOUT", errors.IsTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test message")
			if err.Code != tt.expectedCode {
				t.Errorf("expected code '%s', got '%s'", tt.expectedCode, err.Code)
			}
			if !tt.checkFunc(err) {
				t.Errorf("expected error to be identified as %s", tt.name)
			}
		})
	}
}
