//go:build integration

package audit

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/pkg/postgres"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("EDU_TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("EDU_TEST_POSTGRES_URL not set")
	}
	pool, err := postgres.Connect(context.Background(), postgres.Config{URL: url})
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestPgRecorder_Record(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	recorder := NewPgRecorder(pool)

	actor := "admin-" + uuid.NewString()
	corrID := uuid.NewString()
	entry := Entry{
		ActorUserID:   actor,
		Method:        "POST",
		Path:          "/api/v1/questions",
		StatusCode:    201,
		CorrelationID: corrID,
	}

	if err := recorder.Record(ctx, entry); err != nil {
		t.Fatalf("record: %v", err)
	}

	var (
		gotAction string
		gotMethod string
		gotPath   string
		gotStatus int
		gotCorr   string
	)
	const q = `
		SELECT action, method, path, status_code, correlation_id
		FROM admin_audit_log
		WHERE actor_user_id = $1`
	err := pool.QueryRow(ctx, q, actor).
		Scan(&gotAction, &gotMethod, &gotPath, &gotStatus, &gotCorr)
	if err != nil {
		t.Fatalf("query back: %v", err)
	}

	if gotAction != "POST /api/v1/questions" {
		t.Fatalf("action = %q, want \"POST /api/v1/questions\"", gotAction)
	}
	if gotMethod != "POST" {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/questions" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotStatus != 201 {
		t.Fatalf("status = %d, want 201", gotStatus)
	}
	if gotCorr != corrID {
		t.Fatalf("correlation id = %q, want %q", gotCorr, corrID)
	}
}
