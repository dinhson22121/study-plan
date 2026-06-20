package domain

import (
	"errors"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func seqID() func() string {
	n := 0
	return func() string { n++; return "m" + string(rune('0'+n)) }
}

func topicList(n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = "t" + string(rune('A'+i))
	}
	return out
}

func TestGenerateMilestones_EvenDistribution(t *testing.T) {
	start := time.Unix(1_700_000_000, 0).UTC()
	ms := GenerateMilestones(topicList(10), 4, start, seqID())

	// ceil(10/4)=3 per week -> chunks of 3,3,3,1 = 4 milestones
	if len(ms) != 4 {
		t.Fatalf("expected 4 milestones, got %d", len(ms))
	}
	total := 0
	for i, m := range ms {
		if m.WeekNumber != i+1 {
			t.Fatalf("week %d has WeekNumber %d", i+1, m.WeekNumber)
		}
		want := start.AddDate(0, 0, 7*(i+1))
		if !m.DueDate.Equal(want) {
			t.Fatalf("milestone %d due %v, want %v", i+1, m.DueDate, want)
		}
		total += len(m.TopicIDs)
	}
	if total != 10 {
		t.Fatalf("expected all 10 topics distributed, got %d", total)
	}
	if len(ms[3].TopicIDs) != 1 {
		t.Fatalf("expected last milestone to have 1 topic, got %d", len(ms[3].TopicIDs))
	}
}

func TestGenerateMilestones_FewerTopicsThanWeeks(t *testing.T) {
	ms := GenerateMilestones(topicList(3), 5, time.Unix(0, 0), seqID())
	if len(ms) != 3 {
		t.Fatalf("expected 3 milestones (one per topic), got %d", len(ms))
	}
}

func TestGenerateMilestones_Empty(t *testing.T) {
	if ms := GenerateMilestones(nil, 4, time.Unix(0, 0), seqID()); ms != nil {
		t.Fatalf("expected nil for no topics, got %d", len(ms))
	}
}

func TestNewStudyPlan_Validation(t *testing.T) {
	now := time.Unix(0, 0)
	ms := []Milestone{{ID: "m1", WeekNumber: 1, TopicIDs: []string{"t1"}}}
	if _, err := NewStudyPlan("p", "", "s", "BEGINNER", now, now, ms, now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty user")
	}
	if _, err := NewStudyPlan("p", "u", "s", "BEGINNER", now, now, nil, now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for no milestones")
	}
	p, err := NewStudyPlan("p", "u", "s", "BEGINNER", now, now, ms, now)
	if err != nil || p.FirstMilestone().WeekNumber != 1 {
		t.Fatalf("expected valid plan with first milestone, got %v", err)
	}
}
