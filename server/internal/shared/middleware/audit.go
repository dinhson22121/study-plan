package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/shared/audit"
)

// auditWriteTimeout bounds the best-effort audit write so a slow or unavailable
// database cannot keep a request goroutine alive indefinitely after the
// response has already been sent.
const auditWriteTimeout = 3 * time.Second

// AuditAdmin records a durable audit-trail entry for successful mutating actions
// performed by ADMIN users.
//
// Ordering / gin-context lifecycle:
//
// This middleware is registered once on the /api/v1 group, so in the handler
// chain it runs BEFORE each route's per-route Auth middleware. Auth is what
// calls c.Set(ctxRole, ...). Because we do all our work AFTER c.Next() returns,
// the per-route Auth has already executed by then and the values it stored in
// the *gin.Context persist (gin uses a single Context per request whose Keys map
// is shared across the whole chain). Therefore RoleFrom(c)/UserIDFrom(c) are
// reliably readable here even though this middleware sits "outside" Auth. This
// is the intended post-Next global approach and avoids having to wire the
// middleware into every admin route group individually.
//
// The write is best-effort: any error is logged via zap and never surfaced to
// the client. We use context.Background() (not the request context) for the DB
// write because the request context is cancelled once the handler returns; we
// cap it with a short timeout instead.
func AuditAdmin(recorder audit.Recorder, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if !shouldAudit(c) {
			return
		}

		entry := audit.Entry{
			ActorUserID:   UserIDFrom(c),
			Method:        c.Request.Method,
			Path:          auditPath(c),
			StatusCode:    c.Writer.Status(),
			CorrelationID: CorrelationIDFrom(c),
		}

		// Copy values out of the gin context above; do not reference c or its
		// request context past this point.
		ctx, cancel := context.WithTimeout(context.Background(), auditWriteTimeout)
		defer cancel()
		if err := recorder.Record(ctx, entry); err != nil {
			log.Error("admin_audit_write_failed",
				zap.String("correlation_id", entry.CorrelationID),
				zap.String("actor_user_id", entry.ActorUserID),
				zap.String("action", entry.Action()),
				zap.Int("status", entry.StatusCode),
				zap.Error(err),
			)
		}
	}
}

// shouldAudit reports whether the completed request qualifies for an audit
// entry: an ADMIN actor, a mutating method, and a 2xx response.
func shouldAudit(c *gin.Context) bool {
	if RoleFrom(c) != RoleAdmin {
		return false
	}
	if !isMutatingMethod(c.Request.Method) {
		return false
	}
	status := c.Writer.Status()
	return status >= 200 && status < 300
}

func isMutatingMethod(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

// auditPath prefers the matched route template (c.FullPath, e.g.
// /api/v1/questions/:id) to keep stored paths bounded and free of high-
// cardinality identifiers; it falls back to the raw request path when no route
// matched.
func auditPath(c *gin.Context) string {
	if p := c.FullPath(); p != "" {
		return p
	}
	return c.Request.URL.Path
}
