// Package domain defines the content bounded context: Lessons (attached to a
// curriculum topic) and their ContentItems, plus the repository port.
package domain

import (
	"strings"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// ContentKind is a value object for the type of a content item.
type ContentKind string

const (
	KindPDF   ContentKind = "PDF"
	KindSlide ContentKind = "SLIDE"
	KindNote  ContentKind = "NOTE"
	KindVideo ContentKind = "VIDEO"
)

// Valid reports whether k is a known content kind.
func (k ContentKind) Valid() bool {
	switch k {
	case KindPDF, KindSlide, KindNote, KindVideo:
		return true
	default:
		return false
	}
}

// requiresURL reports whether this kind is referenced by URL (vs inline body).
func (k ContentKind) requiresURL() bool { return k != KindNote }

// ContentItem is one piece of learning material within a lesson. URL-backed
// kinds (PDF/SLIDE/VIDEO) carry a URL; NOTE carries inline Body text.
type ContentItem struct {
	ID         string
	Kind       ContentKind
	URL        string
	Body       string
	OrderIndex int
}

// Lesson is the aggregate root: a titled collection of content items for a topic.
type Lesson struct {
	ID         string
	TopicID    string
	Title      string
	OrderIndex int
	Items      []ContentItem
}

// NewLesson validates and constructs a lesson with its items.
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
