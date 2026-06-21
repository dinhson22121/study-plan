package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/son-ngo/edu-app/internal/curriculum/domain"
)

type fakeCache struct {
	store map[string][]byte
}

func newFakeCache() *fakeCache { return &fakeCache{store: map[string][]byte{}} }

func (c *fakeCache) GetJSON(_ context.Context, key string, dest any) (bool, error) {
	b, ok := c.store[key]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(b, dest)
}
func (c *fakeCache) SetJSON(_ context.Context, key string, val any, _ time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.store[key] = b
	return nil
}
func (c *fakeCache) Delete(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(c.store, k)
	}
	return nil
}

func TestService_CacheAside_ServesAndInvalidates(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, WithCache(newFakeCache(), time.Minute))
	ctx := context.Background()

	if _, err := svc.CreateSubject(ctx, "MATH", "Toán", 10); err != nil {
		t.Fatalf("create A: %v", err)
	}
	if got, _ := svc.ListSubjects(ctx); len(got) != 1 {
		t.Fatalf("first list: want 1, got %d", len(got))
	}

	// Inject a subject straight into the repo (bypassing the service's
	// invalidation) — a correct cache must still serve the stale list.
	b, _ := domain.NewSubject("b", "ENG", "Tiếng Anh", 10)
	_ = repo.CreateSubject(ctx, b)
	if got, _ := svc.ListSubjects(ctx); len(got) != 1 {
		t.Fatalf("second list should be served from cache (1), got %d", len(got))
	}

	// A write through the service invalidates the cache → fresh read sees all 3.
	if _, err := svc.CreateSubject(ctx, "PHY", "Vật Lý", 10); err != nil {
		t.Fatalf("create C: %v", err)
	}
	if got, _ := svc.ListSubjects(ctx); len(got) != 3 {
		t.Fatalf("after invalidation want 3, got %d", len(got))
	}
}

func TestService_NoCache_AlwaysHitsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo) // no cache
	ctx := context.Background()

	if _, err := svc.CreateSubject(ctx, "MATH", "Toán", 10); err != nil {
		t.Fatalf("create: %v", err)
	}
	b, _ := domain.NewSubject("b", "ENG", "Tiếng Anh", 10)
	_ = repo.CreateSubject(ctx, b)
	if got, _ := svc.ListSubjects(ctx); len(got) != 2 {
		t.Fatalf("without cache want 2, got %d", len(got))
	}
}
