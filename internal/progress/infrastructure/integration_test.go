//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/progress/domain"
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

func seedUserTopic(t *testing.T, pool *pgxpool.Pool) (userID, topicID string) {
	t.Helper()
	ctx := context.Background()
	userID = uuid.NewString()
	subjectID, chapterID, tID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	ex := func(sql string, args ...any) {
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	ex(`INSERT INTO users (id,email,display_name,created_at,updated_at) VALUES ($1,$2,'T',NOW(),NOW())`, userID, userID+"@x.com")
	ex(`INSERT INTO subject (id,code,name,grade_level) VALUES ($1,$2,'S',12)`, subjectID, "S-"+subjectID[:8])
	ex(`INSERT INTO chapter (id,subject_id,title,order_index) VALUES ($1,$2,'C',0)`, chapterID, subjectID)
	ex(`INSERT INTO topic (id,chapter_id,title,order_index) VALUES ($1,$2,'Tp',0)`, tID, chapterID)
	return userID, tID
}

func TestProgressRepo_MasteryStreakAchievement(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	userID, topicID := seedUserTopic(t, pool)

	p := domain.TopicProgress{UserID: userID, TopicID: topicID, Status: domain.StatusCompleted, BestScore: 90, Attempts: 2, UpdatedAt: time.Now()}
	if err := repo.UpsertTopicProgress(ctx, &p); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}
	got, err := repo.GetTopicProgress(ctx, userID, topicID)
	if err != nil || got.BestScore != 90 {
		t.Fatalf("get progress: %+v / %v", got, err)
	}
	list, err := repo.ListProgressByUser(ctx, userID)
	if err != nil || len(list) != 1 {
		t.Fatalf("list progress: %d / %v", len(list), err)
	}

	s := domain.Streak{UserID: userID, CurrentStreak: 3, LongestStreak: 5, LastActiveDate: time.Now()}
	if err := repo.UpsertStreak(ctx, &s); err != nil {
		t.Fatalf("upsert streak: %v", err)
	}
	gotS, err := repo.GetStreak(ctx, userID)
	if err != nil || gotS.CurrentStreak != 3 {
		t.Fatalf("get streak: %+v / %v", gotS, err)
	}

	a := &domain.Achievement{UserID: userID, Type: domain.AchievementTopicCompleted, Ref: topicID, UnlockedAt: time.Now()}
	if err := repo.SaveAchievement(ctx, a); err != nil {
		t.Fatalf("save achievement: %v", err)
	}
	has, err := repo.HasAchievement(ctx, userID, domain.AchievementTopicCompleted, topicID)
	if err != nil || !has {
		t.Fatalf("expected achievement present: %v / %v", has, err)
	}
	if err := repo.SaveAchievement(ctx, a); err != nil {
		t.Fatalf("re-save (ON CONFLICT DO NOTHING) should not error: %v", err)
	}
}
