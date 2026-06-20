//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

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
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func seedUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO users (id,email,display_name,created_at,updated_at) VALUES ($1,$2,'T',NOW(),NOW())`,
		id, id+"@x.com"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func TestActivityRepo_InactiveDetection(t *testing.T) {
	pool := testPool(t)
	repo := NewPgActivityRepo(pool)
	ctx := context.Background()

	active := seedUser(t, pool)
	stale := seedUser(t, pool)
	now := time.Now()

	if err := repo.Append(ctx, active, now); err != nil {
		t.Fatalf("append active: %v", err)
	}
	if err := repo.Append(ctx, stale, now.AddDate(0, 0, -10)); err != nil {
		t.Fatalf("append stale: %v", err)
	}

	inactive, err := repo.InactiveUserIDs(ctx, now.AddDate(0, 0, -3))
	if err != nil {
		t.Fatalf("inactive: %v", err)
	}

	set := map[string]bool{}
	for _, id := range inactive {
		set[id] = true
	}
	if set[active] {
		t.Fatalf("recently-active user must not be flagged inactive")
	}
	if !set[stale] {
		t.Fatalf("stale user (10 days) should be flagged inactive")
	}
}
