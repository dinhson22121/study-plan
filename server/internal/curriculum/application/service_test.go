package application

import (
	"context"
	"errors"
	"testing"

	"github.com/son-ngo/edu-app/internal/curriculum/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	subjects map[string]*domain.Subject
	chapters map[string]*domain.Chapter
	topics   map[string]*domain.Topic
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		subjects: map[string]*domain.Subject{},
		chapters: map[string]*domain.Chapter{},
		topics:   map[string]*domain.Topic{},
	}
}

func (r *fakeRepo) CreateSubject(_ context.Context, s *domain.Subject) error {
	r.subjects[s.ID] = s
	return nil
}
func (r *fakeRepo) ListSubjects(context.Context) ([]domain.Subject, error) {
	var out []domain.Subject
	for _, s := range r.subjects {
		out = append(out, *s)
	}
	return out, nil
}
func (r *fakeRepo) GetSubject(_ context.Context, id string) (*domain.Subject, error) {
	if s, ok := r.subjects[id]; ok {
		return s, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) CreateChapter(_ context.Context, c *domain.Chapter) error {
	r.chapters[c.ID] = c
	return nil
}
func (r *fakeRepo) ListChaptersBySubject(_ context.Context, sid string) ([]domain.Chapter, error) {
	var out []domain.Chapter
	for _, c := range r.chapters {
		if c.SubjectID == sid {
			out = append(out, *c)
		}
	}
	return out, nil
}
func (r *fakeRepo) GetChapter(_ context.Context, id string) (*domain.Chapter, error) {
	if c, ok := r.chapters[id]; ok {
		return c, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) CreateTopic(_ context.Context, t *domain.Topic) error {
	r.topics[t.ID] = t
	return nil
}
func (r *fakeRepo) ListTopicsByChapter(_ context.Context, cid string) ([]domain.Topic, error) {
	var out []domain.Topic
	for _, t := range r.topics {
		if t.ChapterID == cid {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (r *fakeRepo) ListTopicsBySubject(_ context.Context, subjectID string) ([]domain.Topic, error) {
	var out []domain.Topic
	for _, t := range r.topics {
		if c, ok := r.chapters[t.ChapterID]; ok && c.SubjectID == subjectID {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (r *fakeRepo) GetTopic(_ context.Context, id string) (*domain.Topic, error) {
	if t, ok := r.topics[id]; ok {
		return t, nil
	}
	return nil, shared.ErrNotFound
}

func TestCreateChapter_RequiresExistingSubject(t *testing.T) {
	svc := NewService(newFakeRepo())
	if _, err := svc.CreateChapter(context.Background(), "ghost", "Ch1", 0); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing subject, got %v", err)
	}
}

func TestCreateTopic_RequiresExistingChapter(t *testing.T) {
	svc := NewService(newFakeRepo())
	if _, err := svc.CreateTopic(context.Background(), "ghost", "T1", 0); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing chapter, got %v", err)
	}
}

func TestCreateHierarchy_HappyPath(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()

	subject, err := svc.CreateSubject(ctx, "MATH", "Toán", 12)
	if err != nil {
		t.Fatalf("create subject: %v", err)
	}
	chapter, err := svc.CreateChapter(ctx, subject.ID, "Logarit", 1)
	if err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	topic, err := svc.CreateTopic(ctx, chapter.ID, "Khái niệm Log", 1)
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}

	got, err := svc.GetTopic(ctx, topic.ID)
	if err != nil || got.Title != "Khái niệm Log" {
		t.Fatalf("get topic mismatch: %+v / %v", got, err)
	}
	topics, _ := svc.ListTopics(ctx, chapter.ID)
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}
}

func TestCreateSubject_ValidationPropagates(t *testing.T) {
	svc := NewService(newFakeRepo())
	if _, err := svc.CreateSubject(context.Background(), "", "n", 12); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
