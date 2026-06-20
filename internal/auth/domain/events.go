package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// EventUserRegistered is the stable name of the registration event other modules
// subscribe to (user creates a profile; notification seeds preferences).
const EventUserRegistered = "auth.user_registered"

// UserRegisteredEvent is published when a new account is created. It carries the
// minimum other contexts need to react, with no auth internals leaked.
type UserRegisteredEvent struct {
	shared.BaseEvent
	UserID string
	Email  string
	Role   Role
}

// NewUserRegisteredEvent builds the event with timestamp and aggregate id set.
func NewUserRegisteredEvent(userID, email string, role Role, at time.Time) UserRegisteredEvent {
	return UserRegisteredEvent{
		BaseEvent: shared.BaseEvent{ID: userID, OccurredAtV: at, AggregateV: userID},
		UserID:    userID,
		Email:     email,
		Role:      role,
	}
}

// EventName satisfies shared.DomainEvent.
func (UserRegisteredEvent) EventName() string { return EventUserRegistered }
