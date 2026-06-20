//go:build integration

// Integration tests for the question Postgres repository. Run with:
//
//	make migrate-up && go test -tags=integration ./internal/question/...
//
// Requires EDU_TEST_POSTGRES_URL; skips if unset.
package infrastructure

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/question/domain"
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

// seedTopic inserts a subject -> chapter -> topic chain via raw SQL and returns
// the topic id, satisfying question's FK.
func seedTopic(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	subjectID, chapterID, topicID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	mustExec(t, pool, `INSERT INTO subject (id, code, name, grade_level) VALUES ($1,$2,$3,12)`,
		subjectID, "S-"+subjectID[:8], "Subj")
	mustExec(t, pool, `INSERT INTO chapter (id, subject_id, title, order_index) VALUES ($1,$2,'Ch',0)`,
		chapterID, subjectID)
	mustExec(t, pool, `INSERT INTO topic (id, chapter_id, title, order_index) VALUES ($1,$2,'Tp',0)`,
		topicID, chapterID)
	return topicID
}

func mustExec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("seed exec: %v", err)
	}
}

func TestQuestionRepo_CreateGetList(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	topicID := seedTopic(t, pool)

	q, _ := domain.NewQuestion(uuid.NewString(), topicID, domain.TypeMCQ, "2+2?", domain.DifficultyEasy, "math",
		[]domain.AnswerOption{
			{ID: uuid.NewString(), Text: "3", OrderIndex: 0},
			{ID: uuid.NewString(), Text: "4", IsCorrect: true, OrderIndex: 1},
		})
	if err := repo.Create(ctx, q); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, q.ID)
	if err != nil || len(got.Options) != 2 || !got.IsCorrect(q.Options[1].ID) {
		t.Fatalf("get: %+v / %v", got, err)
	}

	list, err := repo.List(ctx, domain.ListFilter{TopicID: topicID, Difficulty: domain.DifficultyEasy, Limit: 10})
	if err != nil || len(list) != 1 || len(list[0].Options) != 2 {
		t.Fatalf("list: %d / %v", len(list), err)
	}
}

func TestQuestionRepo_InvalidTopicIsValidationError(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)

	q, _ := domain.NewQuestion(uuid.NewString(), uuid.NewString(), domain.TypeFreeText, "explain", domain.DifficultyHard, "", nil)
	err := repo.Create(context.Background(), q)
	if !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for missing topic FK, got %v", err)
	}
}
