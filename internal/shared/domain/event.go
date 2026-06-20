package domain

import "time"

// DomainEvent is the contract every cross-module event implements. Events are
// the only sanctioned way modules communicate — no module imports another
// module's structs directly.
//
// EventName is a stable type discriminator (e.g. "user.registered"). OccurredAt
// records when the event happened. AggregateID identifies the entity the event
// is about, enabling routing and correlation.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() string
}

// BaseEvent is an embeddable helper that satisfies the timestamp/identity parts
// of DomainEvent. Concrete events embed it and add EventName plus their payload.
type BaseEvent struct {
	ID          string
	OccurredAtV time.Time
	AggregateV  string
}

func (b BaseEvent) OccurredAt() time.Time { return b.OccurredAtV }
func (b BaseEvent) AggregateID() string   { return b.AggregateV }
