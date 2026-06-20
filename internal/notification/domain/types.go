package domain

import shared "github.com/son-ngo/edu-app/internal/shared/domain"

type NotificationType string

const (
	TypeDailyReminder  NotificationType = "DAILY_REMINDER"
	TypeWeeklyQuiz     NotificationType = "WEEKLY_QUIZ"
	TypeStudyPlan      NotificationType = "STUDY_PLAN"
	TypeAchievement    NotificationType = "ACHIEVEMENT"
	TypeReengagement   NotificationType = "REENGAGEMENT"
	TypeAdminBroadcast NotificationType = "ADMIN_BROADCAST"
)

func AllTypes() []NotificationType {
	return []NotificationType{
		TypeDailyReminder, TypeWeeklyQuiz, TypeStudyPlan,
		TypeAchievement, TypeReengagement, TypeAdminBroadcast,
	}
}

func (t NotificationType) Valid() bool {
	switch t {
	case TypeDailyReminder, TypeWeeklyQuiz, TypeStudyPlan,
		TypeAchievement, TypeReengagement, TypeAdminBroadcast:
		return true
	default:
		return false
	}
}

func ParseType(s string) (NotificationType, error) {
	t := NotificationType(s)
	if !t.Valid() {
		return "", shared.ErrValidation.WithMessage("unknown notification type: " + s)
	}
	return t, nil
}

type NotificationStatus string

const (
	StatusPending  NotificationStatus = "PENDING"
	StatusSent     NotificationStatus = "SENT"
	StatusFailed   NotificationStatus = "FAILED"
	StatusRetrying NotificationStatus = "RETRYING"
	StatusSkipped  NotificationStatus = "SKIPPED"
)

type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
)

func (p Platform) Valid() bool { return p == PlatformAndroid || p == PlatformIOS }
