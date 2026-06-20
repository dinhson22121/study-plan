//go:build integration

// Integration tests for the studyplan Postgres repository. Run with:
//
//	make migrate-up && go test -tags=integration ./internal/studyplan/...
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

	"github.com/son-ngo/edu-app/internal/studyplan/domain"
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

// seedTopics creates user, subject, chapter and n topics; returns user id,
// subject id, and topic ids (satisfying studyplan FKs).
func seedTopics(t *testing.T, pool *pgxpool.Pool, n int) (userID, subjectID string, topicIDs []string) {
	t.Helper()
	ctx := context.Background()
	userID, subjectID = uuid.NewString(), uuid.NewString()
	chapterID := uuid.NewString()
	ex := func(sql string, args ...any) {
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	ex(`INSERT INTO users (id,email,display_name,created_at,updated_at) VALUES ($1,$2,'T',NOW(),NOW())`, userID, userID+"@x.com")
	ex(`INSERT INTO subject (id,code,name,grade_level) VALUES ($1,$2,'S',12)`, subjectID, "S-"+subjectID[:8])
	ex(`INSERT INTO chapter (id,subject_id,title,order_index) VALUES ($1,$2,'C',0)`, chapterID, subjectID)
	for i := 0; i < n; i++ {
		tid := uuid.NewString()
		ex(`INSERT INTO topic (id,chapter_id,title,order_index) VALUES ($1,$2,'Tp',$3)`, tid, chapterID, i)
		topicIDs = append(topicIDs, tid)
	}
	return userID, subjectID, topicIDs
}

func TestStudyPlanRepo_SaveAndLoad(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	userID, subjectID, topicIDs := seedTopics(t, pool, 4)

	now := time.Now()
	milestones := domain.GenerateMilestones(topicIDs, 2, now, uuid.NewString)
	plan, err := domain.NewStudyPlan(uuid.NewString(), userID, subjectID, "BEGINNER", now, now.Add(14*24*time.Hour), milestones, now)
	if err != nil {
		t.Fatalf("new plan: %v", err)
	}
	if err := repo.Save(ctx, plan); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.GetByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Milestones) != len(milestones) {
		t.Fatalf("expected %d milestones, got %d", len(milestones), len(got.Milestones))
	}
	total := 0
	for _, m := range got.Milestones {
		total += len(m.TopicIDs)
	}
	if total != 4 {
		t.Fatalf("expected 4 topics across milestones, got %d", total)
	}

	plans, err := repo.ListByUser(ctx, userID)
	if err != nil || len(plans) != 1 {
		t.Fatalf("list: %d / %v", len(plans), err)
	}
}
