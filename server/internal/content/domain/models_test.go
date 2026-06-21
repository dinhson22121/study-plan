package domain

import (
	"errors"
	"testing"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestNewLesson_HappyPath(t *testing.T) {
	items := []ContentItem{
		{ID: "i1", Kind: KindPDF, URL: "https://x/p.pdf"},
		{ID: "i2", Kind: KindNote, Body: "Ghi chú Logarit"},
	}
	l, err := NewLesson("l1", "t1", " Logarit cơ bản ", 1, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Title != "Logarit cơ bản" || len(l.Items) != 2 {
		t.Fatalf("lesson built wrong: %+v", l)
	}
}

func TestNewLesson_Validation(t *testing.T) {
	if _, err := NewLesson("l1", "", "t", 0, nil); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for missing topic id")
	}
	if _, err := NewLesson("l1", "t1", "", 0, nil); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty title")
	}
}

func TestNewLesson_ItemValidation(t *testing.T) {
	cases := []struct {
		name string
		item ContentItem
	}{
		{"bad kind", ContentItem{ID: "i", Kind: ContentKind("GIF")}},
		{"pdf without url", ContentItem{ID: "i", Kind: KindPDF}},
		{"video without url", ContentItem{ID: "i", Kind: KindVideo}},
		{"note without body", ContentItem{ID: "i", Kind: KindNote}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewLesson("l1", "t1", "title", 0, []ContentItem{tc.item}); !errors.Is(err, shared.ErrValidation) {
				t.Fatalf("expected validation error")
			}
		})
	}
}
