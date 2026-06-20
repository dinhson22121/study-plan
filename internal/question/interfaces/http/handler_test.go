package questionhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/question/application"
	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

type stubRepo struct{ q *domain.Question }

func (r *stubRepo) Create(context.Context, *domain.Question) error { return nil }
func (r *stubRepo) GetByID(_ context.Context, id string) (*domain.Question, error) {
	if r.q != nil && r.q.ID == id {
		return r.q, nil
	}
	return nil, shared.ErrNotFound
}
func (r *stubRepo) List(context.Context, domain.ListFilter) ([]domain.Question, error) {
	return []domain.Question{*r.q}, nil
}

func sampleQuestion() *domain.Question {
	q, _ := domain.NewQuestion("q1", "t1", domain.TypeMCQ, "2+2?", domain.DifficultyEasy, "because math",
		[]domain.AnswerOption{{ID: "o1", Text: "3"}, {ID: "o2", Text: "4", IsCorrect: true}})
	return q
}

func newRouter(role string) *gin.Engine {
	svc := application.NewService(&stubRepo{q: sampleQuestion()})
	validate := func(token string) (*middleware.Claims, error) {
		return &middleware.Claims{UserID: "u1", Role: role}, nil
	}
	r := gin.New()
	NewHandler(svc, validate).Routes(r.Group("/api/v1"))
	return r
}

func getQuestion(role string) *httptest.ResponseRecorder {
	r := newRouter(role)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/questions/q1", nil)
	req.Header.Set("Authorization", "Bearer t")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetQuestion_StudentDoesNotSeeAnswers(t *testing.T) {
	w := getQuestion(middleware.RoleStudent)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	body := w.Body.String()
	if contains(body, "is_correct") {
		t.Fatalf("student response leaked is_correct: %s", body)
	}
	if contains(body, "because math") {
		t.Fatalf("student response leaked explanation: %s", body)
	}
}

func TestGetQuestion_AdminSeesAnswers(t *testing.T) {
	w := getQuestion(middleware.RoleAdmin)
	var env struct {
		Data struct {
			Explanation string `json:"explanation"`
			Options     []struct {
				IsCorrect *bool `json:"is_correct"`
			} `json:"options"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if env.Data.Explanation != "because math" {
		t.Fatalf("admin should see explanation, got %q", env.Data.Explanation)
	}
	seen := false
	for _, o := range env.Data.Options {
		if o.IsCorrect != nil {
			seen = true
		}
	}
	if !seen {
		t.Fatalf("admin should see is_correct flags")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
