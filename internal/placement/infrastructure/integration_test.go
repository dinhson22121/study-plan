//go:build integration

package infrastructure

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/placement/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
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

func seedQuestions(t *testing.T, pool *pgxpool.Pool, n int) (userID, subjectID string, qids []string) {
	t.Helper()
	ctx := context.Background()
	userID, subjectID = uuid.NewString(), uuid.NewString()
	chapterID, topicID := uuid.NewString(), uuid.NewString()
	ex := func(sql string, args ...any) {
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	ex(`INSERT INTO users (id,email,display_name,created_at,updated_at) VALUES ($1,$2,'T',NOW(),NOW())`, userID, userID+"@x.com")
	ex(`INSERT INTO subject (id,code,name,grade_level) VALUES ($1,$2,'S',12)`, subjectID, "S-"+subjectID[:8])
	ex(`INSERT INTO chapter (id,subject_id,title,order_index) VALUES ($1,$2,'C',0)`, chapterID, subjectID)
	ex(`INSERT INTO topic (id,chapter_id,title,order_index) VALUES ($1,$2,'Tp',0)`, topicID, chapterID)
	for i := 0; i < n; i++ {
		qid := uuid.NewString()
		ex(`INSERT INTO question (id,topic_id,type,stem,difficulty,explanation) VALUES ($1,$2,'MCQ','q','EASY','')`, qid, topicID)
		qids = append(qids, qid)
	}
	return userID, subjectID, qids
}

func TestPlacementRepo_TestAndResultLifecycle(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	userID, subjectID, qids := seedQuestions(t, pool, 3)

	test, _ := domain.NewPlacementTest(uuid.NewString(), userID, subjectID, qids, time.Now())
	if err := repo.SaveTest(ctx, test); err != nil {
		t.Fatalf("save test: %v", err)
	}
	got, err := repo.GetTest(ctx, test.ID)
	if err != nil || len(got.QuestionIDs) != 3 {
		t.Fatalf("get test: %+v / %v", got, err)
	}
	res := &domain.PlacementResult{ID: uuid.NewString(), UserID: userID, SubjectID: subjectID, Score: 66.67, Level: domain.LevelIntermediate, CompletedAt: time.Now()}
	if err := repo.CompleteWithResult(ctx, test.ID, res); err != nil {
		t.Fatalf("complete with result: %v", err)
	}
	got, _ = repo.GetTest(ctx, test.ID)
	if got.Status != domain.StatusCompleted {
		t.Fatalf("expected COMPLETED, got %s", got.Status)
	}
	latest, err := repo.LatestResult(ctx, userID, subjectID)
	if err != nil || latest.Level != domain.LevelIntermediate {
		t.Fatalf("latest result: %+v / %v", latest, err)
	}

	if _, err := repo.LatestResult(ctx, userID, uuid.NewString()); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected not found for unknown subject, got %v", err)
	}
}
