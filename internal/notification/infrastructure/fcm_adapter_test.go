package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// fakeSender is a programmable pushSender.
type fakeSender struct {
	errs    []error // returned per attempt, in order; nil/short means success
	invalid bool    // whether IsTokenInvalid reports true
	calls   int
}

func (f *fakeSender) Send(context.Context, string, string, string, map[string]string) error {
	var err error
	if f.calls < len(f.errs) {
		err = f.errs[f.calls]
	}
	f.calls++
	return err
}
func (f *fakeSender) IsTokenInvalid(err error) bool { return err != nil && f.invalid }

// stubRepo embeds the interface (nil) and overrides only DeactivateToken, which
// is the sole method the adapter calls. Any other call would panic, surfacing
// an unexpected dependency.
type stubRepo struct {
	domain.Repository
	deactivated []string
}

func (r *stubRepo) DeactivateToken(_ context.Context, token string) error {
	r.deactivated = append(r.deactivated, token)
	return nil
}

// newTestAdapter builds an adapter with sleeping disabled for fast tests.
func newTestAdapter(sender pushSender, repo domain.Repository) *FCMAdapter {
	a := NewFCMAdapter(sender, repo, zap.NewNop())
	a.sleep = func(time.Duration) {}
	return a
}

func TestFCMAdapter_SuccessFirstTry(t *testing.T) {
	s := &fakeSender{}
	a := newTestAdapter(s, &stubRepo{})
	if err := a.Send(context.Background(), "tok", "t", "b", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.calls != 1 {
		t.Fatalf("expected 1 call, got %d", s.calls)
	}
}

func TestFCMAdapter_RetriesThenSucceeds(t *testing.T) {
	s := &fakeSender{errs: []error{errors.New("transient"), errors.New("transient")}} // 3rd attempt succeeds
	a := newTestAdapter(s, &stubRepo{})
	if err := a.Send(context.Background(), "tok", "t", "b", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.calls != 3 {
		t.Fatalf("expected 3 calls (2 fail, 1 ok), got %d", s.calls)
	}
}

func TestFCMAdapter_MaxRetriesExceeded(t *testing.T) {
	s := &fakeSender{errs: []error{errors.New("t"), errors.New("t"), errors.New("t")}}
	a := newTestAdapter(s, &stubRepo{})
	err := a.Send(context.Background(), "tok", "t", "b", nil)
	if !errors.Is(err, shared.ErrMaxRetriesExceeded) {
		t.Fatalf("expected ErrMaxRetriesExceeded, got %v", err)
	}
	if s.calls != 3 {
		t.Fatalf("expected exactly 3 attempts, got %d", s.calls)
	}
}

func TestFCMAdapter_InvalidTokenDeactivatesNoRetry(t *testing.T) {
	s := &fakeSender{errs: []error{errors.New("unregistered")}, invalid: true}
	repo := &stubRepo{}
	a := newTestAdapter(s, repo)

	err := a.Send(context.Background(), "bad-token", "t", "b", nil)
	if !errors.Is(err, shared.ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
	if s.calls != 1 {
		t.Fatalf("invalid token must not retry; got %d calls", s.calls)
	}
	if len(repo.deactivated) != 1 || repo.deactivated[0] != "bad-token" {
		t.Fatalf("expected token to be deactivated, got %v", repo.deactivated)
	}
}
