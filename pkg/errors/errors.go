// Package errors defines custom error types and utilities for error handling in SSH Arena.
package errors

import (
	"fmt"
)

// ErrorCode represents a machine‑readable error code.
type ErrorCode string

const (
	// ErrCodeNotFound indicates a requested resource was not found.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeInvalidInput indicates invalid user input or command.
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
	// ErrCodeUnauthorized indicates insufficient permissions.
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// ErrCodeInternal indicates an unexpected internal error.
	ErrCodeInternal ErrorCode = "INTERNAL"
	// ErrCodeConflict indicates a conflict with current state (e.g., duplicate).
	ErrCodeConflict ErrorCode = "CONFLICT"
	// ErrCodeTimeout indicates an operation timed out.
	ErrCodeTimeout ErrorCode = "TIMEOUT"
	// ErrCodeUnavailable indicates a service is temporarily unavailable.
	ErrCodeUnavailable ErrorCode = "UNAVAILABLE"
)

// GameError is a structured error that includes a code, a message, and an underlying cause.
type GameError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error returns the error message, satisfying the error interface.
func (e *GameError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause, if any.
func (e *GameError) Unwrap() error {
	return e.Cause
}

// New creates a new GameError with the given code and message.
func New(code ErrorCode, message string) *GameError {
	return &GameError{Code: code, Message: message}
}

// Wrap wraps an existing error with a code and message.
func Wrap(code ErrorCode, message string, cause error) *GameError {
	return &GameError{Code: code, Message: message, Cause: cause}
}

// IsNotFound checks if the error is a GameError with code ErrCodeNotFound.
func IsNotFound(err error) bool {
	if e, ok := err.(*GameError); ok {
		return e.Code == ErrCodeNotFound
	}
	return false
}

// IsInvalidInput checks if the error is a GameError with code ErrCodeInvalidInput.
func IsInvalidInput(err error) bool {
	if e, ok := err.(*GameError); ok {
		return e.Code == ErrCodeInvalidInput
	}
	return false
}

// IsUnauthorized checks if the error is a GameError with code ErrCodeUnauthorized.
func IsUnauthorized(err error) bool {
	if e, ok := err.(*GameError); ok {
		return e.Code == ErrCodeUnauthorized
	}
	return false
}

// Helper functions to create common errors.

// NotFound returns a GameError with ErrCodeNotFound.
func NotFound(resource string) *GameError {
	return New(ErrCodeNotFound, fmt.Sprintf("resource %q not found", resource))
}

// InvalidInput returns a GameError with ErrCodeInvalidInput.
func InvalidInput(details string) *GameError {
	return New(ErrCodeInvalidInput, fmt.Sprintf("invalid input: %s", details))
}

// Unauthorized returns a GameError with ErrCodeUnauthorized.
func Unauthorized(action string) *GameError {
	return New(ErrCodeUnauthorized, fmt.Sprintf("unauthorized to perform %s", action))
}

// Internal returns a GameError with ErrCodeInternal.
func Internal(reason string) *GameError {
	return New(ErrCodeInternal, fmt.Sprintf("internal error: %s", reason))
}

// Conflict returns a GameError with ErrCodeConflict.
func Conflict(resource string) *GameError {
	return New(ErrCodeConflict, fmt.Sprintf("conflict with resource %q", resource))
}

// Timeout returns a GameError with ErrCodeTimeout.
func Timeout(operation string) *GameError {
	return New(ErrCodeTimeout, fmt.Sprintf("operation %q timed out", operation))
}

// Unavailable returns a GameError with ErrCodeUnavailable.
func Unavailable(service string) *GameError {
	return New(ErrCodeUnavailable, fmt.Sprintf("service %q is unavailable", service))
}