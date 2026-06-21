// Package audit records admin actions to a durable audit trail. The Recorder
// interface is intentionally tiny (one method) so callers — chiefly the
// AuditAdmin middleware — can depend on the abstraction and tests can supply a
// fake. Implementations live alongside this file (see pg.go).
package audit

import (
	"context"
	"time"
)

// Entry is a single audit-trail record describing one admin action.
type Entry struct {
	ActorUserID   string
	Method        string
	Path          string
	StatusCode    int
	CorrelationID string
	CreatedAt     time.Time
}

// Action returns the canonical "METHOD path" action string stored alongside the
// structured columns for quick human scanning.
func (e Entry) Action() string {
	return e.Method + " " + e.Path
}

// Recorder persists audit entries. Implementations must be safe for concurrent
// use. Record returns an error so callers can decide how to handle failures;
// the middleware treats writes as best-effort and only logs on error.
type Recorder interface {
	Record(ctx context.Context, entry Entry) error
}
