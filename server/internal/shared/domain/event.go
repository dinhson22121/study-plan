package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() string
}

type BaseEvent struct {
	ID          string
	OccurredAtV time.Time
	AggregateV  string
}

func (b BaseEvent) OccurredAt() time.Time { return b.OccurredAtV }
func (b BaseEvent) AggregateID() string   { return b.AggregateV }
