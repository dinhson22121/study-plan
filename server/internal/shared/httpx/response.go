package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

func List(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: &meta})
}

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

func StatusForCode(code string) int {
	switch code {
	case domain.ErrNotFound.Code:
		return http.StatusNotFound
	case domain.ErrValidation.Code:
		return http.StatusUnprocessableEntity
	case domain.ErrUnauthorized.Code:
		return http.StatusUnauthorized
	case domain.ErrForbidden.Code, domain.ErrPreferenceDisabled.Code:
		return http.StatusForbidden
	case domain.ErrConflict.Code:
		return http.StatusConflict
	case domain.ErrTokenInvalid.Code:
		return http.StatusGone
	case domain.ErrMaxRetriesExceeded.Code:
		return http.StatusServiceUnavailable
	case domain.ErrDuplicateMessage.Code:
		return http.StatusOK
	case domain.ErrTooManyRequests.Code:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
