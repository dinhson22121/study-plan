package domain

import "time"

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
