package httpx

import (
	"errors"
	"net/http"
	"testing"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestStatusForCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not found", domain.ErrNotFound, http.StatusNotFound},
		{"validation", domain.ErrValidation, http.StatusUnprocessableEntity},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden},
		{"pref disabled", domain.ErrPreferenceDisabled, http.StatusForbidden},
		{"conflict", domain.ErrConflict, http.StatusConflict},
		{"fcm token invalid", domain.ErrTokenInvalid, http.StatusGone},
		{"max retries", domain.ErrMaxRetriesExceeded, http.StatusServiceUnavailable},
		{"duplicate idempotent", domain.ErrDuplicateMessage, http.StatusOK},
		{"unknown -> 500", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			de := domain.AsDomainError(tc.err)
			if got := StatusForCode(de.Code); got != tc.want {
				t.Fatalf("StatusForCode(%s) = %d, want %d", de.Code, got, tc.want)
			}
		})
	}
}
