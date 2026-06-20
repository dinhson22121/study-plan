package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestDomainError_IsMatchesByCode(t *testing.T) {
	// Arrange: a sentinel wrapped with extra context.
	wrapped := ErrNotFound.WithCause(errors.New("row missing"))

	// Act & Assert: errors.Is should match the sentinel by Code.
	if !errors.Is(wrapped, ErrNotFound) {
		t.Fatalf("expected wrapped error to match ErrNotFound")
	}
	if errors.Is(wrapped, ErrConflict) {
		t.Fatalf("did not expect wrapped error to match ErrConflict")
	}
}

func TestDomainError_WithCauseDoesNotMutateSentinel(t *testing.T) {
	// Arrange
	cause := errors.New("boom")

	// Act
	derived := ErrInternal.WithCause(cause)

	// Assert: sentinel stays clean, derived carries the cause.
	if ErrInternal.Err != nil {
		t.Fatalf("sentinel ErrInternal was mutated: %v", ErrInternal.Err)
	}
	if !errors.Is(derived, cause) {
		t.Fatalf("derived error should wrap the cause")
	}
}

func TestAsDomainError_WrapsUnknownAsInternal(t *testing.T) {
	// Arrange
	plain := fmt.Errorf("some lib error")

	// Act
	de := AsDomainError(plain)

	// Assert
	if de.Code != ErrInternal.Code {
		t.Fatalf("expected INTERNAL code, got %s", de.Code)
	}
	if !errors.Is(de, plain) {
		t.Fatalf("expected wrapped original error to be retrievable")
	}
}

func TestAsDomainError_PassesThroughDomainError(t *testing.T) {
	if got := AsDomainError(ErrValidation); got.Code != ErrValidation.Code {
		t.Fatalf("expected VALIDATION_ERROR, got %s", got.Code)
	}
	if AsDomainError(nil) != nil {
		t.Fatalf("expected nil for nil input")
	}
}
