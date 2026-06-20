//go:build integration

// Integration tests for the goal Postgres repository. Run with:
//
//	make migrate-up && go test -tags=integration ./internal/goal/...
//
// Requires EDU_TEST_POSTGRES_URL; skips if unset.
package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/goal/domain"
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

func seedUserAndSubject(t *testing.T, pool *pgxpool.Pool) (userID, subjectID string) {
	t.Helper()
	ctx := context.Background()
	userID, subjectID = uuid.NewString(), uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, email, display_name, created_at, updated_at) VALUES ($1,$2,'T',NOW(),NOW())`,
		userID, userID+"@x.com"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO subject (id, code, name, grade_level) VALUES ($1,$2,'Subj',12)`,
		subjectID, "S-"+subjectID[:8]); err != nil {
		t.Fatalf("seed subject: %v", err)
	}
	return userID, subjectID
}

func TestGoalRepo_UpsertAndGet(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	userID, subjectID := seedUserAndSubject(t, pool)

	g, _ := domain.NewGoal(userID, "HUST", "CNTT", time.Now().Add(60*24*time.Hour), 2, 5,
		[]domain.SubjectTarget{{SubjectID: subjectID, CurrentScore: 5, TargetScore: 8}}, time.Now())
	if err := repo.Upsert(ctx, g); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := repo.GetByUserID(ctx, userID)
	if err != nil || got.TargetUniversity != "HUST" || len(got.Subjects) != 1 {
		t.Fatalf("get: %+v / %v", got, err)
	}

	// Upsert again replaces.
	g.TargetUniversity = "VNU"
	if err := repo.Upsert(ctx, g); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	got, _ = repo.GetByUserID(ctx, userID)
	if got.TargetUniversity != "VNU" || len(got.Subjects) != 1 {
		t.Fatalf("expected replaced goal, got %+v", got)
	}
}
