package domain

import (
	"errors"
	"testing"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestNewSubject(t *testing.T) {
	s, err := NewSubject("id", " math ", " Toán ", 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Code != "MATH" || s.Name != "Toán" {
		t.Fatalf("subject not normalized: %+v", s)
	}

	cases := []struct {
		name  string
		code  string
		sname string
		grade int
	}{
		{"empty code", "", "n", 12},
		{"empty name", "M", "", 12},
		{"grade too low", "M", "n", 0},
		{"grade too high", "M", "n", 13},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewSubject("id", tc.code, tc.sname, tc.grade); !errors.Is(err, shared.ErrValidation) {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestNewChapterAndTopic(t *testing.T) {
	if _, err := NewChapter("id", "", "t", 0); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for missing subject id")
	}
	if _, err := NewChapter("id", "s", "", 0); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty title")
	}
	if _, err := NewChapter("id", "s", "t", -1); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for negative order")
	}
	if _, err := NewTopic("id", "", "t", 0); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for missing chapter id")
	}
	c, err := NewChapter("id", "s", " Logarit ", 2)
	if err != nil || c.Title != "Logarit" {
		t.Fatalf("chapter not built/trimmed: %+v / %v", c, err)
	}
}
