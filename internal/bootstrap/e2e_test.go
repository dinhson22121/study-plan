//go:build integration

// End-to-end test driving the full student activity loop through the real HTTP
// router over a real Postgres + Redis. Run with:
//
//	make migrate-up && go test -tags=integration ./internal/bootstrap/...
//
// Requires EDU_TEST_POSTGRES_URL and EDU_TEST_REDIS_URL. Kafka is not required:
// achievement/reminder pushes go through Kafka best-effort and are allowed to
// fail here; the synchronous in-process eventbus drives progress/analytics, which
// is what this test asserts. Skips if the env vars are unset.
package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/config"
	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
	"github.com/son-ngo/edu-app/pkg/kafka"
	"github.com/son-ngo/edu-app/pkg/postgres"
	"github.com/son-ngo/edu-app/pkg/redis"
)

func e2eDeps(t *testing.T) *app.Deps {
	t.Helper()
	pgURL := os.Getenv("EDU_TEST_POSTGRES_URL")
	redisURL := os.Getenv("EDU_TEST_REDIS_URL")
	if pgURL == "" || redisURL == "" {
		t.Skip("EDU_TEST_POSTGRES_URL / EDU_TEST_REDIS_URL not set")
	}
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	db, err := postgres.Connect(ctx, postgres.Config{URL: pgURL})
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	t.Cleanup(db.Close)
	rdb, err := redis.Connect(ctx, redisURL)
	if err != nil {
		t.Fatalf("redis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })

	cfg := &config.Config{Env: "test", Timezone: "UTC"}
	cfg.JWT.Secret = "e2e-secret"
	cfg.JWT.AccessTTL = time.Hour
	cfg.JWT.RefreshTTL = time.Hour
	cfg.JWT.Issuer = "edu-app-e2e"
	cfg.Kafka.GroupID = "e2e"
	cfg.Kafka.Partitions = 1

	kc := kafka.NewClient([]string{"localhost:9092"}) // not started; pushes are best-effort
	return &app.Deps{
		Cfg: cfg, DB: db, Redis: rdb, Kafka: kc, Producer: kc.NewProducer(),
		Bus: eventbus.New(), Log: zap.NewNop(),
	}
}

// seedCatalog inserts a subject/chapter/topic with n single-correct MCQs and
// returns the subject id, topic id, and a map of question id -> correct option id.
func seedCatalog(t *testing.T, deps *app.Deps, n int) (subjectID, topicID string, correct map[string]string) {
	t.Helper()
	ctx := context.Background()
	subjectID, topicID = uuid.NewString(), uuid.NewString()
	chapterID := uuid.NewString()
	correct = map[string]string{}
	ex := func(sql string, args ...any) {
		if _, err := deps.DB.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	ex(`INSERT INTO subject (id,code,name,grade_level) VALUES ($1,$2,'Toán',12)`, subjectID, "MATH-"+subjectID[:8])
	ex(`INSERT INTO chapter (id,subject_id,title,order_index) VALUES ($1,$2,'Logarit',0)`, chapterID, subjectID)
	ex(`INSERT INTO topic (id,chapter_id,title,order_index) VALUES ($1,$2,'Khái niệm Log',0)`, topicID, chapterID)
	for i := 0; i < n; i++ {
		qid, right, wrong := uuid.NewString(), uuid.NewString(), uuid.NewString()
		ex(`INSERT INTO question (id,topic_id,type,stem,difficulty,explanation) VALUES ($1,$2,'MCQ','q?','EASY','vì vậy')`, qid, topicID)
		ex(`INSERT INTO answer_option (id,question_id,text,is_correct,order_index) VALUES ($1,$2,'đúng',true,0)`, right, qid)
		ex(`INSERT INTO answer_option (id,question_id,text,is_correct,order_index) VALUES ($1,$2,'sai',false,1)`, wrong, qid)
		correct[qid] = right
	}
	return subjectID, topicID, correct
}

// client issues authenticated JSON requests against the router.
type client struct {
	t      *testing.T
	router *gin.Engine
	token  string
}

func (c *client) do(method, path string, body any) map[string]any {
	c.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)
	if w.Code >= 400 {
		c.t.Fatalf("%s %s -> %d: %s", method, path, w.Code, w.Body.String())
	}
	var env map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	return env
}

func TestE2E_StudentActivityLoop(t *testing.T) {
	deps := e2eDeps(t)
	router, _ := BuildRouter(deps)
	subjectID, topicID, correct := seedCatalog(t, deps, 3)

	c := &client{t: t, router: router}

	// Register a student (also seeds profile + notification preferences via events).
	email := uuid.NewString() + "@e2e.test"
	reg := c.do(http.MethodPost, "/api/v1/auth/register", gin.H{"email": email, "password": "password1"})
	c.token = reg["data"].(map[string]any)["access_token"].(string)
	if c.token == "" {
		t.Fatal("no access token from register")
	}

	// Set a goal, then generate a study plan for the subject.
	c.do(http.MethodPut, "/api/v1/goals", gin.H{
		"target_university": "HUST", "target_major": "CNTT",
		"target_date":   time.Now().Add(60 * 24 * time.Hour).Format(time.RFC3339),
		"hours_per_day": 2, "days_per_week": 5,
		"subjects": []gin.H{{"subject_id": subjectID, "current_score": 5, "target_score": 8}},
	})
	plan := c.do(http.MethodPost, "/api/v1/studyplans/generate", gin.H{"subject_id": subjectID})
	if ms, _ := plan["data"].(map[string]any)["milestones"].([]any); len(ms) == 0 {
		t.Fatalf("expected study plan milestones, got %v", plan["data"])
	}

	// Take a quiz on the topic and answer everything correctly.
	start := c.do(http.MethodPost, "/api/v1/quizzes", gin.H{"topic_id": topicID})
	data := start["data"].(map[string]any)
	quizID := data["id"].(string)
	var answers []gin.H
	for _, q := range data["question_ids"].([]any) {
		qid := q.(string)
		answers = append(answers, gin.H{"question_id": qid, "option_id": correct[qid]})
	}
	res := c.do(http.MethodPost, "/api/v1/quizzes/"+quizID+"/submit", gin.H{"answers": answers})
	if score := res["data"].(map[string]any)["score"].(float64); score != 100 {
		t.Fatalf("expected score 100, got %v", score)
	}

	// Progress should reflect the mastered topic and a streak of 1.
	prog := c.do(http.MethodGet, "/api/v1/progress", nil)
	pd := prog["data"].(map[string]any)
	if pd["topics_completed"].(float64) != 1 || pd["current_streak"].(float64) != 1 {
		t.Fatalf("unexpected progress: %v", pd)
	}

	// Analytics dashboard should show the quiz average and completion.
	dash := c.do(http.MethodGet, "/api/v1/analytics/me", nil)
	dd := dash["data"].(map[string]any)
	if dd["quiz_average"].(float64) != 100 || dd["topics_completed"].(float64) != 1 {
		t.Fatalf("unexpected dashboard: %v", dd)
	}
}
