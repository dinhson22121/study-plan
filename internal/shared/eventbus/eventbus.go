package eventbus

import (
	"context"
	"errors"
	"sync"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

type Handler func(ctx context.Context, evt domain.DomainEvent) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func New() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

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
