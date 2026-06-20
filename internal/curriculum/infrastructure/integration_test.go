//go:build integration

// Integration tests for the curriculum Postgres repository. Run with:
//
//	make migrate-up && go test -tags=integration ./internal/curriculum/...
//
// Requires EDU_TEST_POSTGRES_URL; skips if unset.
package infrastructure

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/curriculum/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/pkg/postgres"
)

func testRepo(t *testing.T) *PgRepository {
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
	return NewPgRepository(pool)
}

func TestCurriculumHierarchy(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()

	subject, _ := domain.NewSubject(uuid.NewString(), "MATH-"+uuid.NewString()[:8], "Toán", 12)
	if err := repo.CreateSubject(ctx, subject); err != nil {
		t.Fatalf("create subject: %v", err)
	}
	chapter, _ := domain.NewChapter(uuid.NewString(), subject.ID, "Logarit", 1)
	if err := repo.CreateChapter(ctx, chapter); err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	topic, _ := domain.NewTopic(uuid.NewString(), chapter.ID, "Khái niệm Log", 1)
	if err := repo.CreateTopic(ctx, topic); err != nil {
		t.Fatalf("create topic: %v", err)
	}

	got, err := repo.GetTopic(ctx, topic.ID)
	if err != nil || got.Title != "Khái niệm Log" {
		t.Fatalf("get topic: %+v / %v", got, err)
	}
	topics, err := repo.ListTopicsByChapter(ctx, chapter.ID)
	if err != nil || len(topics) != 1 {
		t.Fatalf("list topics: %d / %v", len(topics), err)
	}
	chapters, err := repo.ListChaptersBySubject(ctx, subject.ID)
	if err != nil || len(chapters) != 1 {
		t.Fatalf("list chapters: %d / %v", len(chapters), err)
	}

	if _, err := repo.GetSubject(ctx, uuid.NewString()); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected not found for missing subject, got %v", err)
	}
}
