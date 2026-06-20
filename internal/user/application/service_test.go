package application

import (
	"context"
	"errors"
	"testing"
	"time"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	userdomain "github.com/son-ngo/edu-app/internal/user/domain"
)

type fakeUserRepo struct {
	users    map[string]*userdomain.User
	failNext error
}

func newFakeUserRepo() *fakeUserRepo { return &fakeUserRepo{users: map[string]*userdomain.User{}} }

func (r *fakeUserRepo) Create(_ context.Context, u *userdomain.User) error {
	if r.failNext != nil {
		err := r.failNext
		r.failNext = nil
		return err
	}
	if _, ok := r.users[u.ID]; ok {
		return shared.ErrConflict
	}
	cp := *u
	r.users[u.ID] = &cp
	return nil
}
func (r *fakeUserRepo) FindByID(_ context.Context, id string) (*userdomain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeUserRepo) Update(_ context.Context, u *userdomain.User) error {
	if _, ok := r.users[u.ID]; !ok {
		return shared.ErrNotFound
	}
	cp := *u
	r.users[u.ID] = &cp
	return nil
}

func newSvc(repo *fakeUserRepo) *Service {
	return NewService(repo, WithClock(func() time.Time { return time.Unix(1000, 0).UTC() }))
}

func TestHandleUserRegistered_CreatesProfileWithDerivedName(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newSvc(repo)

	evt := authdomain.NewUserRegisteredEvent("u1", "minh@example.com", authdomain.RoleStudent, time.Unix(0, 0))
	if err := svc.HandleUserRegistered(context.Background(), evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	u := repo.users["u1"]
	if u == nil || u.DisplayName != "minh" {
		t.Fatalf("profile not created with derived name: %+v", u)
	}
}

func TestHandleUserRegistered_IsIdempotentOnDuplicate(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newSvc(repo)
	evt := authdomain.NewUserRegisteredEvent("u1", "a@b.com", authdomain.RoleStudent, time.Unix(0, 0))

	_ = svc.HandleUserRegistered(context.Background(), evt)
	if err := svc.HandleUserRegistered(context.Background(), evt); err != nil {
		t.Fatalf("replayed event should be idempotent, got %v", err)
	}
}

func TestHandleUserRegistered_RejectsWrongEventType(t *testing.T) {
	svc := newSvc(newFakeUserRepo())
	if err := svc.HandleUserRegistered(context.Background(), stubEvent{}); !errors.Is(err, shared.ErrInternal) {
		t.Fatalf("expected internal error for wrong event type, got %v", err)
	}
}

func TestUpdateDisplayName(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newSvc(repo)
	evt := authdomain.NewUserRegisteredEvent("u1", "a@b.com", authdomain.RoleStudent, time.Unix(0, 0))
	_ = svc.HandleUserRegistered(context.Background(), evt)

	updated, err := svc.UpdateDisplayName(context.Background(), "u1", "New Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.DisplayName != "New Name" {
		t.Fatalf("display name not updated: %q", updated.DisplayName)
	}
	if _, err := svc.UpdateDisplayName(context.Background(), "u1", "  "); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for blank name, got %v", err)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	svc := newSvc(newFakeUserRepo())
	if _, err := svc.GetProfile(context.Background(), "ghost"); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

// stubEvent is a non-matching DomainEvent for the wrong-type test.
type stubEvent struct{}

func (stubEvent) EventName() string     { return "other.event" }
func (stubEvent) OccurredAt() time.Time { return time.Unix(0, 0) }
func (stubEvent) AggregateID() string   { return "x" }
