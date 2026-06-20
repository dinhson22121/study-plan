//go:build integration

// Integration tests for the content Postgres repository. Run with:
//
//	make migrate-up && go test -tags=integration ./internal/content/...
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

	"github.com/son-ngo/edu-app/internal/content/domain"
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

func seedTopic(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	subjectID, chapterID, topicID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	exec(t, pool, `INSERT INTO subject (id, code, name, grade_level) VALUES ($1,$2,$3,12)`, subjectID, "S-"+subjectID[:8], "Subj")
	exec(t, pool, `INSERT INTO chapter (id, subject_id, title, order_index) VALUES ($1,$2,'Ch',0)`, chapterID, subjectID)
	exec(t, pool, `INSERT INTO topic (id, chapter_id, title, order_index) VALUES ($1,$2,'Tp',0)`, topicID, chapterID)
	return topicID
}

func exec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("seed exec: %v", err)
	}
}

func TestContentRepo_CreateAndList(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	topicID := seedTopic(t, pool)

	lesson, _ := domain.NewLesson(uuid.NewString(), topicID, "Logarit", 0, []domain.ContentItem{
		{ID: uuid.NewString(), Kind: domain.KindPDF, URL: "https://x/p.pdf", OrderIndex: 0},
		{ID: uuid.NewString(), Kind: domain.KindNote, Body: "ghi chú", OrderIndex: 1},
	})
	if err := repo.CreateLesson(ctx, lesson); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetLesson(ctx, lesson.ID)
	if err != nil || len(got.Items) != 2 {
		t.Fatalf("get: %+v / %v", got, err)
	}
	list, err := repo.ListByTopic(ctx, topicID)
	if err != nil || len(list) != 1 || len(list[0].Items) != 2 {
		t.Fatalf("list: %d / %v", len(list), err)
	}
}

func TestContentRepo_InvalidTopicIsValidationError(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)

	lesson, _ := domain.NewLesson(uuid.NewString(), uuid.NewString(), "X", 0, nil)
	err := repo.CreateLesson(context.Background(), lesson)
	if !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for missing topic FK, got %v", err)
	}
}
