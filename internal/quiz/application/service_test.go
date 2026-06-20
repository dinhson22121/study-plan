package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/son-ngo/edu-app/internal/quiz/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	sessions map[string]*domain.QuizSession
	results  map[string]*domain.QuizResult
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{sessions: map[string]*domain.QuizSession{}, results: map[string]*domain.QuizResult{}}
}
func (r *fakeRepo) SaveSession(_ context.Context, s *domain.QuizSession) error {
	cp := *s
	r.sessions[s.ID] = &cp
	return nil
}
func (r *fakeRepo) GetSession(_ context.Context, id string) (*domain.QuizSession, error) {
	if s, ok := r.sessions[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) SaveResultAndComplete(_ context.Context, res *domain.QuizResult) error {
	s, ok := r.sessions[res.SessionID]
	if !ok {
		return shared.ErrNotFound
	}
	s.Status = domain.StatusCompleted
	cp := *res
	r.results[res.SessionID] = &cp
	return nil
}
func (r *fakeRepo) GetResultForUser(_ context.Context, sessionID, userID string) (*domain.QuizResult, error) {
	if res, ok := r.results[sessionID]; ok && res.UserID == userID {
		return res, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) ListResultsByUser(_ context.Context, userID string) ([]domain.QuizResult, error) {
	var out []domain.QuizResult
	for _, res := range r.results {
		if res.UserID == userID {
			out = append(out, *res)
		}
	}
	return out, nil
}

type fakeSource struct {
	ids     []string
	details map[string]domain.QuestionDetail
}

func (s *fakeSource) SampleForTopic(_ context.Context, _ string, limit int) ([]string, error) {
	if limit < len(s.ids) {
		return s.ids[:limit], nil
	}
	return s.ids, nil
}
func (s *fakeSource) Details(_ context.Context, _ []string) (map[string]domain.QuestionDetail, error) {
	return s.details, nil
}

type fakeBus struct{ events []shared.DomainEvent }

func (b *fakeBus) Publish(_ context.Context, e shared.DomainEvent) error {
	b.events = append(b.events, e)
	return nil
}

func newService(repo *fakeRepo, src *fakeSource, bus *fakeBus) *Service {
	s := NewService(repo, src, bus)
	s.now = func() time.Time { return time.Unix(1000, 0).UTC() }
	seq := 0
	s.newID = func() string { seq++; return "id" + string(rune('0'+seq)) }
	return s
}

func source2() *fakeSource {
	return &fakeSource{
		ids: []string{"q1", "q2"},
		details: map[string]domain.QuestionDetail{
			"q1": {CorrectOptionIDs: []string{"a"}, Explanation: "e1"},
			"q2": {CorrectOptionIDs: []string{"b"}, Explanation: "e2"},
		},
	}
}

func TestStartQuiz_AssemblesSession(t *testing.T) {
	repo, src, bus := newFakeRepo(), source2(), &fakeBus{}
	svc := newService(repo, src, bus)

	session, err := svc.StartQuiz(context.Background(), "u1", "t1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(session.QuestionIDs) != 2 || session.Status != domain.StatusInProgress {
		t.Fatalf("session not assembled: %+v", session)
	}
}

func TestSubmitQuiz_GradesPublishesAndCompletes(t *testing.T) {
	repo, src, bus := newFakeRepo(), source2(), &fakeBus{}
	svc := newService(repo, src, bus)
	session, _ := svc.StartQuiz(context.Background(), "u1", "t1", 2)

	result, err := svc.SubmitQuiz(context.Background(), session.ID, "u1", []AnswerInput{
		{QuestionID: "q1", OptionID: "a"},
		{QuestionID: "q2", OptionID: "b"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score != 100.0 || !result.Passed {
		t.Fatalf("expected 100%% pass, got %.1f", result.Score)
	}
	if len(result.Reviews) != 2 || result.Reviews[0].Explanation != "e1" {
		t.Fatalf("expected review revealed, got %+v", result.Reviews)
	}
	if repo.sessions[session.ID].Status != domain.StatusCompleted {
		t.Fatalf("session not marked completed")
	}
	if len(bus.events) != 1 || bus.events[0].EventName() != domain.EventQuizCompleted {
		t.Fatalf("expected quiz.completed event, got %+v", bus.events)
	}
}

func TestSubmitQuiz_Guards(t *testing.T) {
	repo, src, bus := newFakeRepo(), source2(), &fakeBus{}
	svc := newService(repo, src, bus)
	session, _ := svc.StartQuiz(context.Background(), "u1", "t1", 2)

	if _, err := svc.SubmitQuiz(context.Background(), session.ID, "intruder", nil); !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	_, _ = svc.SubmitQuiz(context.Background(), session.ID, "u1", []AnswerInput{{QuestionID: "q1", OptionID: "a"}})
	if _, err := svc.SubmitQuiz(context.Background(), session.ID, "u1", nil); !errors.Is(err, shared.ErrConflict) {
		t.Fatalf("expected conflict on double submit, got %v", err)
	}
}
