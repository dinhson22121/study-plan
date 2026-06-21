package domain

import (
	"errors"
	"fmt"
)

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

func (e *DomainError) Unwrap() error { return e.Err }

func (e *DomainError) Is(target error) bool {
	var t *DomainError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

func (e *DomainError) WithCause(cause error) *DomainError {
	return &DomainError{Code: e.Code, Message: e.Message, Err: cause}
}

func (e *DomainError) WithMessage(msg string) *DomainError {
	return &DomainError{Code: e.Code, Message: msg, Err: e.Err}
}

func newDomain(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

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
	ErrTooManyRequests    = newDomain("TOO_MANY_REQUESTS", "rate limit exceeded, retry later")
)

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
