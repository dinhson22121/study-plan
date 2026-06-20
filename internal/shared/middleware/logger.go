package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const correlationHeader = "X-Correlation-ID"

const ctxCorrelationID = "correlation_id"

const maxCorrelationIDLen = 64

func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		corrID := c.GetHeader(correlationHeader)
		if corrID == "" || len(corrID) > maxCorrelationIDLen {
			corrID = uuid.NewString()
		}
		c.Set(ctxCorrelationID, corrID)
		c.Header(correlationHeader, corrID)

		c.Next()

		log.Info("http_request",
			zap.String("correlation_id", corrID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic_recovered",
					zap.String("correlation_id", CorrelationIDFrom(c)),
					zap.Any("panic", r),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

func CorrelationIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxCorrelationID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
