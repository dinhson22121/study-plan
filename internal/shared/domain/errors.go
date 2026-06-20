// Package domain holds cross-cutting domain primitives shared by every module:
// the canonical error type, sentinel errors, and the domain-event contract.
package domain

import (
	"errors"
	"fmt"
)

// DomainError is the canonical error type raised by domain and application
// layers. Code is a stable machine-readable identifier (mapped to HTTP status
// at the interface boundary); Message is a human-readable summary; Err is an
// optional wrapped cause used for logging and errors.Is/As traversal.
type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the wrapped cause for errors.Is / errors.As.
func (e *DomainError) Unwrap() error { return e.Err }

// Is reports equality by Code so sentinel comparison survives wrapping:
// errors.Is(wrapped, ErrNotFound) holds when the wrapped error carries the
// same Code, regardless of Message or cause.
func (e *DomainError) Is(target error) bool {
	var t *DomainError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// WithCause returns a copy of the error carrying an underlying cause without
// mutating the shared sentinel (sentinels must stay immutable).
func (e *DomainError) WithCause(cause error) *DomainError {
	return &DomainError{Code: e.Code, Message: e.Message, Err: cause}
}

// WithMessage returns a copy of the error with a more specific message while
// preserving the Code used for HTTP mapping and Is comparison.
func (e *DomainError) WithMessage(msg string) *DomainError {
	return &DomainError{Code: e.Code, Message: msg, Err: e.Err}
}

func newDomain(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

// Sentinel errors. Cross-cutting ones live here; modules may declare their own
// DomainError sentinels in their own packages.
var (
	ErrNotFound           = newDomain("NOT_FOUND", "resource not found")
	ErrValidation         = newDomain("VALIDATION_ERROR", "request failed validation")
	ErrUnauthorized       = newDomain("UNAUTHORIZED", "authentication required or invalid")
	ErrForbidden          = newDomain("FORBIDDEN", "not permitted to access this resource")
	ErrConflict           = newDomain("CONFLICT", "resource already exists or state conflict")
	ErrInternal           = newDomain("INTERNAL", "unexpected internal error")
	ErrTokenInvalid       = newDomain("FCM_TOKEN_INVALID", "FCM token revoked or expired")
	ErrMaxRetriesExceeded = newDomain("MAX_RETRIES_EXCEEDED", "notification failed after retries")
	ErrPreferenceDisabled = newDomain("NOTIF_DISABLED", "user disabled this notification type")
	ErrDuplicateMessage   = newDomain("DUPLICATE_MESSAGE", "idempotency key already processed")
)

// AsDomainError extracts a *DomainError from any error, returning ErrInternal
// (wrapping the original) when the error is not already a DomainError.
func AsDomainError(err error) *DomainError {
	if err == nil {
		return nil
	}
	var de *DomainError
	if errors.As(err, &de) {
		return de
	}
	return ErrInternal.WithCause(err)
}
