//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/quiz/domain"
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

func seedQuestions(t *testing.T, pool *pgxpool.Pool, n int) (userID, topicID string, qids, oids []string) {
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
	for i := 0; i < n; i++ {
		qid, oid := uuid.NewString(), uuid.NewString()
		ex(`INSERT INTO question (id,topic_id,type,stem,difficulty,explanation) VALUES ($1,$2,'MCQ','q','EASY','')`, qid, tID)
		ex(`INSERT INTO answer_option (id,question_id,text,is_correct,order_index) VALUES ($1,$2,'A',true,0)`, oid, qid)
		qids = append(qids, qid)
		oids = append(oids, oid)
	}
	return userID, tID, qids, oids
}

func TestQuizRepo_Lifecycle(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	userID, topicID, qids, oids := seedQuestions(t, pool, 2)

	session, _ := domain.NewQuizSession(uuid.NewString(), userID, topicID, qids, time.Now())
	if err := repo.SaveSession(ctx, session); err != nil {
		t.Fatalf("save session: %v", err)
	}
	got, err := repo.GetSession(ctx, session.ID)
	if err != nil || len(got.QuestionIDs) != 2 {
		t.Fatalf("get session: %+v / %v", got, err)
	}

	result := domain.QuizResult{
		SessionID: session.ID, UserID: userID, TopicID: topicID,
		Score: 100, CorrectCount: 2, Total: 2, Passed: true, CompletedAt: time.Now(),
		Reviews: []domain.QuestionReview{
			{QuestionID: qids[0], SelectedOptionID: oids[0], IsCorrect: true},
			{QuestionID: qids[1], SelectedOptionID: "", IsCorrect: false},
		},
	}
	if err := repo.SaveResultAndComplete(ctx, &result); err != nil {
		t.Fatalf("save result and complete: %v", err)
	}
	if s, _ := repo.GetSession(ctx, session.ID); s.Status != domain.StatusCompleted {
		t.Fatalf("session should be COMPLETED after save")
	}

	gotRes, err := repo.GetResultForUser(ctx, session.ID, userID)
	if err != nil || gotRes.Score != 100 || len(gotRes.Reviews) != 2 {
		t.Fatalf("get result: %+v / %v", gotRes, err)
	}
	if _, err := repo.GetResultForUser(ctx, session.ID, uuid.NewString()); err == nil {
		t.Fatalf("expected not found for other user's result")
	}
	list, err := repo.ListResultsByUser(ctx, userID)
	if err != nil || len(list) != 1 {
		t.Fatalf("list results: %d / %v", len(list), err)
	}
}
