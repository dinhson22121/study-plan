package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeDraftRepo struct {
	byID     map[string]*domain.QuestionDraft
	byAsset  map[string][]string
	optionBy map[string]*domain.QuestionDraftOption
}

func newFakeDraftRepo() *fakeDraftRepo {
	return &fakeDraftRepo{
		byID:     map[string]*domain.QuestionDraft{},
		byAsset:  map[string][]string{},
		optionBy: map[string]*domain.QuestionDraftOption{},
	}
}

func (r *fakeDraftRepo) addDraft(d *domain.QuestionDraft) {
	cp := *d
	cp.Options = append([]domain.QuestionDraftOption(nil), d.Options...)
	r.byID[d.ID] = &cp
	r.byAsset[d.AssetID] = append(r.byAsset[d.AssetID], d.ID)
	for i := range cp.Options {
		opt := cp.Options[i]
		r.optionBy[opt.ID] = &cp.Options[i]
	}
}

func (r *fakeDraftRepo) ListByAsset(_ context.Context, assetID string) ([]domain.QuestionDraft, error) {
	ids := r.byAsset[assetID]
	out := make([]domain.QuestionDraft, 0, len(ids))
	for _, id := range ids {
		cp := *r.byID[id]
		cp.Options = append([]domain.QuestionDraftOption(nil), r.byID[id].Options...)
		out = append(out, cp)
	}
	return out, nil
}

func (r *fakeDraftRepo) GetByID(_ context.Context, id string) (*domain.QuestionDraft, error) {
	d, ok := r.byID[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	cp := *d
	cp.Options = append([]domain.QuestionDraftOption(nil), d.Options...)
	return &cp, nil
}

func (r *fakeDraftRepo) UpdateDraft(_ context.Context, id, stem, explanation string, at time.Time) error {
	d, ok := r.byID[id]
	if !ok {
		return shared.ErrNotFound
	}
	d.Stem = stem
	d.ExplanationRaw = explanation
	d.UpdatedAt = at
	return nil
}

func (r *fakeDraftRepo) UpdateOption(_ context.Context, optionID, text string, isCorrect bool) error {
	o, ok := r.optionBy[optionID]
	if !ok {
		return shared.ErrNotFound
	}
	o.OptionText = text
	o.IsCorrectInferred = isCorrect
	return nil
}

func (r *fakeDraftRepo) MarkPublished(_ context.Context, id, questionID, reviewedBy string, at time.Time) error {
	d, ok := r.byID[id]
	if !ok {
		return shared.ErrNotFound
	}
	d.Status = domain.DraftPublished
	d.PublishedQuestionID = questionID
	d.ReviewedBy = reviewedBy
	d.ReviewedAt = &at
	d.UpdatedAt = at
	return nil
}

func makeDraft(id, assetID string, status domain.DraftStatus) *domain.QuestionDraft {
	return &domain.QuestionDraft{
		ID:              id,
		AssetID:         assetID,
		QuestionNumber:  1,
		QuestionType:    "MCQ",
		Stem:            "2+2=?",
		ExplanationRaw:  "math",
		Status:          status,
		ParseConfidence: 0.9,
		Options: []domain.QuestionDraftOption{
			{ID: id + "-a", QuestionDraftID: id, OptionLabel: "A", OptionText: "3", OrderIndex: 0},
			{ID: id + "-b", QuestionDraftID: id, OptionLabel: "B", OptionText: "4", IsCorrectInferred: true, OrderIndex: 1},
		},
	}
}

func TestPublishByAsset_PublishesAllPendingDrafts(t *testing.T) {
	drafts := newFakeDraftRepo()
	drafts.addDraft(makeDraft("d1", "asset-1", domain.DraftPending))
	second := makeDraft("d2", "asset-1", domain.DraftPending)
	second.QuestionNumber = 2
	second.Stem = "3+3=?"
	drafts.addDraft(second)

	questions := NewService(newFakeRepo())
	svc := NewDraftService(drafts, questions)
	svc.now = func() time.Time { return time.Unix(123, 0).UTC() }

	got, err := svc.PublishByAsset(context.Background(), PublishByAssetInput{
		AssetID: "asset-1", TopicID: "topic-1", Difficulty: "easy", ReviewedBy: "admin-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 published questions, got %d", len(got))
	}
	if drafts.byID["d1"].Status != domain.DraftPublished || drafts.byID["d2"].Status != domain.DraftPublished {
		t.Fatalf("expected drafts to be marked published")
	}
}

func TestPublishByAsset_RejectsWhenNothingPending(t *testing.T) {
	drafts := newFakeDraftRepo()
	drafts.addDraft(makeDraft("d1", "asset-1", domain.DraftPublished))

	svc := NewDraftService(drafts, NewService(newFakeRepo()))
	_, err := svc.PublishByAsset(context.Background(), PublishByAssetInput{
		AssetID: "asset-1", TopicID: "topic-1", Difficulty: "easy", ReviewedBy: "admin-1",
	})
	if !errors.Is(err, shared.ErrConflict) {
		t.Fatalf("expected conflict when all drafts are published, got %v", err)
	}
}
