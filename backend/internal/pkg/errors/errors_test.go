package errors

import (
	"fmt"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	t.Run("without details", func(t *testing.T) {
		e := &AppError{Code: 1001, Message: "not found"}
		got := e.Error()
		want := "[1001] not found"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("with details", func(t *testing.T) {
		e := &AppError{Code: 1001, Message: "not found", Details: "service 42"}
		got := e.Error()
		want := "[1001] not found: service 42"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("empty details treated as without", func(t *testing.T) {
		e := &AppError{Code: 2001, Message: "bad param", Details: ""}
		got := e.Error()
		want := "[2001] bad param"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestWithDetails(t *testing.T) {
	orig := &AppError{Code: 1001, Message: "not found", Retryable: false}
	got := WithDetails(orig, "service 42")

	// Should return a new copy
	if got == orig {
		t.Error("WithDetails should return a new *AppError, not the same pointer")
	}

	if got.Code != 1001 {
		t.Errorf("Code = %d, want 1001", got.Code)
	}
	if got.Message != "not found" {
		t.Errorf("Message = %q, want %q", got.Message, "not found")
	}
	if got.Details != "service 42" {
		t.Errorf("Details = %q, want %q", got.Details, "service 42")
	}
	if got.Retryable {
		t.Error("Retryable should be false")
	}

	// Original should be unchanged
	if orig.Details != "" {
		t.Errorf("original Details changed to %q", orig.Details)
	}
}

func TestWithDetails_PreservesRetryable(t *testing.T) {
	orig := &AppError{Code: 3001, Message: "workflow failed", Retryable: true}
	got := WithDetails(orig, "step 3 timeout")

	if !got.Retryable {
		t.Error("Retryable should be preserved as true")
	}
}

func TestWrap(t *testing.T) {
	inner := fmt.Errorf("connection refused")
	got := Wrap(ErrInternal, inner)

	if got.Code != ErrInternal.Code {
		t.Errorf("Code = %d, want %d", got.Code, ErrInternal.Code)
	}
	if got.Message != ErrInternal.Message {
		t.Errorf("Message = %q, want %q", got.Message, ErrInternal.Message)
	}
	if got.Details != "connection refused" {
		t.Errorf("Details = %q, want %q", got.Details, "connection refused")
	}
	if got.Retryable {
		t.Error("Retryable should be false for ErrInternal")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"retryable workflow error", ErrWorkflowFailed, true},
		{"retryable external system", ErrExternalSystem, true},
		{"retryable timeout", ErrTimeout, true},
		{"non-retryable not found", ErrNotFound, false},
		{"non-retryable invalid param", ErrInvalidParam, false},
		{"non-retryable internal", ErrInternal, false},
		{"non-retryable unauthorized", ErrUnauthorized, false},
		{"wrapped retryable keeps flag", WithDetails(ErrTimeout, "call timed out"), true},
		{"standard error not retryable", fmt.Errorf("random error"), false},
		{"nil error not retryable", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsAppError(t *testing.T) {
	t.Run("with *AppError", func(t *testing.T) {
		appErr, ok := IsAppError(ErrNotFound)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if appErr.Code != 1001 {
			t.Errorf("Code = %d, want 1001", appErr.Code)
		}
	})

	t.Run("with standard error", func(t *testing.T) {
		_, ok := IsAppError(fmt.Errorf("not an app error"))
		if ok {
			t.Error("expected ok=false for standard error")
		}
	})

	t.Run("with nil", func(t *testing.T) {
		_, ok := IsAppError(nil)
		if ok {
			t.Error("expected ok=false for nil")
		}
	})
}

func TestPredefinedErrors(t *testing.T) {
	// Verify all predefined errors have unique codes
	codes := map[int]string{}
	predefined := []*AppError{
		Success, ErrNotFound, ErrAlreadyExists, ErrConflict,
		ErrInvalidParam, ErrUnauthorized, ErrForbidden,
		ErrWorkflowFailed, ErrExternalSystem, ErrTimeout, ErrInternal,
	}
	for _, e := range predefined {
		if prev, exists := codes[e.Code]; exists {
			t.Errorf("duplicate code %d used by both %q and %q", e.Code, prev, e.Message)
		}
		codes[e.Code] = e.Message
	}
}
