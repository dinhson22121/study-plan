// Package eventbus provides a lightweight in-process publish/subscribe bus for
// domain events. It is used for synchronous, same-process module choreography;
// cross-process / durable delivery goes through Kafka (see pkg/kafka).
//
// Dispatch is synchronous: Publish invokes every subscribed handler inline and
// in registration order. This keeps ordering deterministic and tests simple. A
// handler that needs to do slow or fallible work should hand off to Kafka or
// spawn its own goroutine rather than block the publisher.
package eventbus

import (
	"context"
	"errors"
	"sync"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// Handler reacts to a published domain event. Returning an error does not abort
// dispatch to other handlers; the bus collects errors and returns them joined.
type Handler func(ctx context.Context, evt domain.DomainEvent) error

// Bus is the subscription registry. The zero value is not usable; use New.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// New returns an empty Bus ready for Subscribe/Publish.
func New() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

// Subscribe registers a handler for a specific event name. Multiple handlers may
// subscribe to the same name; they run in registration order on Publish.
func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

// Publish dispatches an event to all handlers subscribed to its EventName.
// Handler errors are aggregated and returned; a panic in one handler does not
// prevent the others from running.
func (b *Bus) Publish(ctx context.Context, evt domain.DomainEvent) error {
	b.mu.RLock()
	hs := make([]Handler, len(b.handlers[evt.EventName()]))
	copy(hs, b.handlers[evt.EventName()])
	b.mu.RUnlock()

	var errs []error
	for _, h := range hs {
		if err := safeInvoke(ctx, h, evt); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func safeInvoke(ctx context.Context, h Handler, evt domain.DomainEvent) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = domain.ErrInternal.WithMessage("event handler panicked")
		}
	}()
	return h(ctx, evt)
}
