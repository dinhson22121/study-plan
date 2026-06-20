package domain

import (
	"strings"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type ContentKind string

const (
	KindPDF   ContentKind = "PDF"
	KindSlide ContentKind = "SLIDE"
	KindNote  ContentKind = "NOTE"
	KindVideo ContentKind = "VIDEO"
)

func (k ContentKind) Valid() bool {
	switch k {
	case KindPDF, KindSlide, KindNote, KindVideo:
		return true
	default:
		return false
	}
}

func (k ContentKind) requiresURL() bool { return k != KindNote }

type ContentItem struct {
	ID         string
	Kind       ContentKind
	URL        string
	Body       string
	OrderIndex int
}

type Lesson struct {
	ID         string
	TopicID    string
	Title      string
	OrderIndex int
	Items      []ContentItem
}

func NewLesson(id, topicID, title string, orderIndex int, items []ContentItem) (*Lesson, error) {
	if topicID == "" {
		return nil, shared.ErrValidation.WithMessage("topic id is required")
	}
	if strings.TrimSpace(title) == "" {
		return nil, shared.ErrValidation.WithMessage("lesson title is required")
	}
	if orderIndex < 0 {
		return nil, shared.ErrValidation.WithMessage("order index must be non-negative")
	}
	for i := range items {
		if err := validateItem(items[i]); err != nil {
			return nil, err
		}
	}
	return &Lesson{
		ID: id, TopicID: topicID, Title: strings.TrimSpace(title),
		OrderIndex: orderIndex, Items: items,
	}, nil
}

func validateItem(it ContentItem) error {
	if !it.Kind.Valid() {
		return shared.ErrValidation.WithMessage("invalid content kind")
	}
	if it.Kind.requiresURL() && strings.TrimSpace(it.URL) == "" {
		return shared.ErrValidation.WithMessage(string(it.Kind) + " item requires a url")
	}
	if it.Kind == KindNote && strings.TrimSpace(it.Body) == "" {
		return shared.ErrValidation.WithMessage("NOTE item requires body text")
	}
	return nil
}
