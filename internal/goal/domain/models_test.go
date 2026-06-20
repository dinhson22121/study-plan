package domain

import (
	"errors"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func validSubjects() []SubjectTarget {
	return []SubjectTarget{{SubjectID: "s1", CurrentScore: 5, TargetScore: 8}}
}

func TestNewGoal_HappyPath(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	target := now.Add(60 * 24 * time.Hour)
	g, err := NewGoal("u1", " HUST ", " CNTT ", target, 2, 5, validSubjects(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.TargetUniversity != "HUST" || g.TargetMajor != "CNTT" {
		t.Fatalf("not trimmed: %+v", g)
	}
}

func TestNewGoal_Validation(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	future := now.Add(30 * 24 * time.Hour)
	cases := []struct {
		name        string
		uid, uni    string
		date        time.Time
		hours, days int
		subjects    []SubjectTarget
	}{
		{"no user", "", "U", future, 2, 5, validSubjects()},
		{"no university", "u1", "", future, 2, 5, validSubjects()},
		{"past date", "u1", "U", now.Add(-time.Hour), 2, 5, validSubjects()},
		{"zero hours", "u1", "U", future, 0, 5, validSubjects()},
		{"days too high", "u1", "U", future, 2, 8, validSubjects()},
		{"no subjects", "u1", "U", future, 2, 5, nil},
		{"bad score", "u1", "U", future, 2, 5, []SubjectTarget{{SubjectID: "s1", CurrentScore: 11}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewGoal(tc.uid, tc.uni, "m", tc.date, tc.hours, tc.days, tc.subjects, now)
			if !errors.Is(err, shared.ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestWeeksUntilTarget(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	g := &Goal{TargetDate: now.Add(21 * 24 * time.Hour)}
	if w := g.WeeksUntilTarget(now); w != 3 {
		t.Fatalf("expected 3 weeks, got %d", w)
	}

	gNear := &Goal{TargetDate: now.Add(2 * 24 * time.Hour)}
	if w := gNear.WeeksUntilTarget(now); w != 1 {
		t.Fatalf("expected clamp to 1 week, got %d", w)
	}
}
