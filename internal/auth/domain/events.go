package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const EventUserRegistered = "auth.user_registered"

type UserRegisteredEvent struct {
	shared.BaseEvent
	UserID string
	Email  string
	Role   Role
}

func NewUserRegisteredEvent(userID, email string, role Role, at time.Time) UserRegisteredEvent {
	return UserRegisteredEvent{
		BaseEvent: shared.BaseEvent{ID: userID, OccurredAtV: at, AggregateV: userID},
		UserID:    userID,
		Email:     email,
		Role:      role,
	}
}

func (UserRegisteredEvent) EventName() string { return EventUserRegistered }
