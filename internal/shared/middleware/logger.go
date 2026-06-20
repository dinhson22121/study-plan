package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// correlationHeader is the inbound/outbound header used to trace a request
// end-to-end across HTTP and Kafka.
const correlationHeader = "X-Correlation-ID"

// ctxCorrelationID is where the per-request correlation id is stashed.
const ctxCorrelationID = "correlation_id"

// Logger returns middleware that assigns/propagates a correlation id and logs
// one structured line per request with method, path, status, and latency.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		corrID := c.GetHeader(correlationHeader)
		if corrID == "" {
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

// Recovery returns middleware that converts a panic into a logged 500 instead of
// crashing the process.
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

// CorrelationIDFrom returns the request's correlation id, or "" if absent.
func CorrelationIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxCorrelationID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
