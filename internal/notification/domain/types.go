// Package domain defines the notification bounded context: device tokens,
// templates, preferences, the delivery log aggregate, the Kafka message
// schemas, and the ports the module depends on.
package domain

import shared "github.com/son-ngo/edu-app/internal/shared/domain"

// NotificationType is a value object identifying a category of notification.
type NotificationType string

const (
	TypeDailyReminder  NotificationType = "DAILY_REMINDER"
	TypeWeeklyQuiz     NotificationType = "WEEKLY_QUIZ"
	TypeStudyPlan      NotificationType = "STUDY_PLAN"
	TypeAchievement    NotificationType = "ACHIEVEMENT"
	TypeReengagement   NotificationType = "REENGAGEMENT"
	TypeAdminBroadcast NotificationType = "ADMIN_BROADCAST"
)

// AllTypes lists every notification type (used to seed default preferences).
func AllTypes() []NotificationType {
	return []NotificationType{
		TypeDailyReminder, TypeWeeklyQuiz, TypeStudyPlan,
		TypeAchievement, TypeReengagement, TypeAdminBroadcast,
	}
}

// Valid reports whether t is a known notification type.
func (t NotificationType) Valid() bool {
	switch t {
	case TypeDailyReminder, TypeWeeklyQuiz, TypeStudyPlan,
		TypeAchievement, TypeReengagement, TypeAdminBroadcast:
		return true
	default:
		return false
	}
}

// ParseType validates and returns a NotificationType.
func ParseType(s string) (NotificationType, error) {
	t := NotificationType(s)
	if !t.Valid() {
		return "", shared.ErrValidation.WithMessage("unknown notification type: " + s)
	}
	return t, nil
}

// NotificationStatus is a value object for the delivery lifecycle state.
type NotificationStatus string

const (
	StatusPending  NotificationStatus = "PENDING"
	StatusSent     NotificationStatus = "SENT"
	StatusFailed   NotificationStatus = "FAILED"
	StatusRetrying NotificationStatus = "RETRYING"
	StatusSkipped  NotificationStatus = "SKIPPED"
)

// Platform is the device OS for a registered token.
type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
)

// Valid reports whether p is a supported platform.
func (p Platform) Valid() bool { return p == PlatformAndroid || p == PlatformIOS }
