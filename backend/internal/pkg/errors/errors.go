package errors

import "fmt"

// AppError is the unified error type for the platform.
type AppError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Retryable bool   `json:"retryable"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Predefined errors
var (
	Success           = &AppError{Code: 0, Message: "success"}
	ErrNotFound       = &AppError{Code: 1001, Message: "resource not found"}
	ErrAlreadyExists  = &AppError{Code: 1002, Message: "resource already exists"}
	ErrConflict       = &AppError{Code: 1003, Message: "resource conflict"}
	ErrInvalidParam   = &AppError{Code: 2001, Message: "invalid parameter"}
	ErrUnauthorized   = &AppError{Code: 2002, Message: "unauthorized"}
	ErrForbidden      = &AppError{Code: 2003, Message: "forbidden"}
	ErrWorkflowFailed = &AppError{Code: 3001, Message: "workflow failed", Retryable: true}
	ErrExternalSystem = &AppError{Code: 3002, Message: "external system error", Retryable: true}
	ErrTimeout        = &AppError{Code: 3003, Message: "operation timeout", Retryable: true}
	ErrInternal       = &AppError{Code: 4001, Message: "internal error"}
)

// WithDetails returns a copy with details appended.
func WithDetails(appErr *AppError, details string) *AppError {
	return &AppError{Code: appErr.Code, Message: appErr.Message, Details: details, Retryable: appErr.Retryable}
}

// Wrap wraps a Go error into an AppError.
func Wrap(appErr *AppError, err error) *AppError {
	return &AppError{Code: appErr.Code, Message: appErr.Message, Details: err.Error(), Retryable: appErr.Retryable}
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable
	}
	return false
}

// IsAppError tries to cast an error to *AppError.
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
