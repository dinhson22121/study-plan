package eventbus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// stubEvent is a minimal DomainEvent for testing the bus.
type stubEvent struct {
	name string
	agg  string
}

func (s stubEvent) EventName() string     { return s.name }
func (s stubEvent) OccurredAt() time.Time { return time.Unix(0, 0) }
func (s stubEvent) AggregateID() string   { return s.agg }

func TestBus_PublishInvokesSubscribedHandlersInOrder(t *testing.T) {
	// Arrange
	bus := New()
	var order []string
	bus.Subscribe("user.registered", func(_ context.Context, _ domain.DomainEvent) error {
		order = append(order, "first")
		return nil
	})
	bus.Subscribe("user.registered", func(_ context.Context, _ domain.DomainEvent) error {
		order = append(order, "second")
		return nil
	})

	// Act
	err := bus.Publish(context.Background(), stubEvent{name: "user.registered", agg: "u1"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Fatalf("handlers ran out of order: %v", order)
	}
}

func TestBus_PublishIgnoresUnrelatedEvents(t *testing.T) {
	bus := New()
	called := false
	bus.Subscribe("a.happened", func(_ context.Context, _ domain.DomainEvent) error {
		called = true
		return nil
	})

	_ = bus.Publish(context.Background(), stubEvent{name: "b.happened"})

	if called {
		t.Fatalf("handler for a.happened should not fire on b.happened")
	}
}

func TestBus_PublishAggregatesHandlerErrors(t *testing.T) {
	bus := New()
	want := errors.New("handler failed")
	bus.Subscribe("x", func(_ context.Context, _ domain.DomainEvent) error { return want })
	bus.Subscribe("x", func(_ context.Context, _ domain.DomainEvent) error { return nil })

	err := bus.Publish(context.Background(), stubEvent{name: "x"})
	if !errors.Is(err, want) {
		t.Fatalf("expected aggregated error to contain handler error, got %v", err)
	}
}

func TestBus_PublishRecoversFromPanic(t *testing.T) {
	bus := New()
	secondRan := false
	bus.Subscribe("p", func(_ context.Context, _ domain.DomainEvent) error { panic("boom") })
	bus.Subscribe("p", func(_ context.Context, _ domain.DomainEvent) error {
		secondRan = true
		return nil
	})

	err := bus.Publish(context.Background(), stubEvent{name: "p"})
	if err == nil {
		t.Fatalf("expected an error from the panicking handler")
	}
	if !secondRan {
		t.Fatalf("a panic in one handler must not stop the others")
	}
}
