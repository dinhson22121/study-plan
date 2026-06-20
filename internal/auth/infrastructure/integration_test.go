//go:build integration

// Integration tests for the auth Postgres/Redis adapters. Run with:
//
//	go test -tags=integration ./internal/auth/infrastructure/...
//
// Requires a reachable Postgres and Redis. Configure via:
//
//	EDU_TEST_POSTGRES_URL=postgres://eduapp:secret@localhost:5432/eduapp?sslmode=disable
//	EDU_TEST_REDIS_URL=redis://localhost:6379/1
//
// Tests skip if those env vars are unset.
package infrastructure

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/pkg/postgres"
	"github.com/son-ngo/edu-app/pkg/redis"
)

func testPostgres(t *testing.T) *PgCredentialRepo {
	t.Helper()
	url := os.Getenv("EDU_TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("EDU_TEST_POSTGRES_URL not set")
	}
	pool, err := postgres.Connect(context.Background(), postgres.Config{URL: url})
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewPgCredentialRepo(pool)
}

func TestPgCredentialRepo_CreateAndFind(t *testing.T) {
	repo := testPostgres(t)
	ctx := context.Background()

	email := uuid.NewString() + "@example.com"
	cred, _ := authdomain.NewUserCredential(uuid.NewString(), email, "hash", authdomain.RoleStudent)
	if err := repo.Create(ctx, cred); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.FindByEmail(ctx, cred.Email)
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	if got.UserID != cred.UserID {
		t.Fatalf("user id mismatch")
	}

	// Duplicate email must conflict.
	dup, _ := authdomain.NewUserCredential(uuid.NewString(), email, "hash", authdomain.RoleStudent)
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict on duplicate email, got %v", err)
	}

	// Unknown lookup must be ErrNotFound.
	if _, err := repo.FindByEmail(ctx, "nobody-"+uuid.NewString()+"@x.com"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestRedisRefreshStore_SaveExistsDelete(t *testing.T) {
	url := os.Getenv("EDU_TEST_REDIS_URL")
	if url == "" {
		t.Skip("EDU_TEST_REDIS_URL not set")
	}
	client, err := redis.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	store := NewRedisRefreshStore(client, time.Minute)
	ctx := context.Background()
	userID, jti := uuid.NewString(), uuid.NewString()

	if err := store.Save(ctx, userID, jti); err != nil {
		t.Fatalf("save: %v", err)
	}
	if ok, _ := store.Exists(ctx, userID, jti); !ok {
		t.Fatalf("expected token to exist after save")
	}
	if err := store.Delete(ctx, userID, jti); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if ok, _ := store.Exists(ctx, userID, jti); ok {
		t.Fatalf("expected token to be gone after delete")
	}
}
