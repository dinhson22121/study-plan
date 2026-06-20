package domain

import "time"

// GenerateMilestones distributes topics sequentially (in curriculum order)
// across the given number of weeks, evenly, one milestone per week. Each
// milestone's due date is the end of its week (start + 7*week days). Weeks is
// clamped to at least 1; empty topic input yields no milestones.
//
// This is the "sequential by curriculum order" strategy chosen for v1. Weak-
// topic prioritization (using placement detail) is a future enhancement.
func GenerateMilestones(topicIDs []string, weeks int, start time.Time, newID func() string) []Milestone {
	n := len(topicIDs)
	if n == 0 {
		return nil
	}
	if weeks < 1 {
		weeks = 1
	}
	perWeek := ceilDiv(n, weeks)

	var milestones []Milestone
	for week := 0; week*perWeek < n; week++ {
		lo := week * perWeek
		hi := lo + perWeek
		if hi > n {
			hi = n
		}
		chunk := make([]string, hi-lo)
		copy(chunk, topicIDs[lo:hi])
		milestones = append(milestones, Milestone{
			ID:         newID(),
			WeekNumber: week + 1,
			TopicIDs:   chunk,
			DueDate:    start.AddDate(0, 0, 7*(week+1)),
		})
	}
	return milestones
}

func ceilDiv(a, b int) int {
	if b <= 0 {
		return a
	}
	return (a + b - 1) / b
}
