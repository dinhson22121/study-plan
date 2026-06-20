// Package httpx holds HTTP-layer helpers shared by every module's handlers:
// a consistent response envelope and the DomainError -> HTTP status mapping.
package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// Envelope is the uniform response shape for every endpoint.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError is the client-facing error payload. Code is the stable DomainError
// code; Message is human-readable. Internal causes are never serialized.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta carries pagination metadata for list responses.
type Meta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// OK writes a 200 success envelope.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

// Created writes a 201 success envelope.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

// List writes a 200 success envelope with pagination metadata.
func List(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: &meta})
}

// Fail maps any error to the correct HTTP status and writes an error envelope.
// DUPLICATE_MESSAGE is intentionally mapped to 200 (idempotent success) per the
// PRD: a replayed idempotency key is not a client error.
func Fail(c *gin.Context, err error) {
	de := domain.AsDomainError(err)
	status := StatusForCode(de.Code)

	if status == http.StatusOK {
		c.JSON(http.StatusOK, Envelope{Success: true})
		return
	}
	c.JSON(status, Envelope{
		Success: false,
		Error:   &APIError{Code: de.Code, Message: de.Message},
	})
}

// StatusForCode maps a DomainError code to an HTTP status (PRD section 9).
func StatusForCode(code string) int {
	switch code {
	case domain.ErrNotFound.Code:
		return http.StatusNotFound // 404
	case domain.ErrValidation.Code:
		return http.StatusUnprocessableEntity // 422
	case domain.ErrUnauthorized.Code:
		return http.StatusUnauthorized // 401
	case domain.ErrForbidden.Code, domain.ErrPreferenceDisabled.Code:
		return http.StatusForbidden // 403
	case domain.ErrConflict.Code:
		return http.StatusConflict // 409
	case domain.ErrTokenInvalid.Code:
		return http.StatusGone // 410
	case domain.ErrMaxRetriesExceeded.Code:
		return http.StatusServiceUnavailable // 503
	case domain.ErrDuplicateMessage.Code:
		return http.StatusOK // 200 idempotent
	default:
		return http.StatusInternalServerError // 500
	}
}
