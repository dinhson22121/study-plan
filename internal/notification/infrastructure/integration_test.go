//go:build integration

// Integration tests for the notification Postgres repository and Redis
// idempotency store. Run with:
//
//	make migrate-up   # ensure schema exists
//	go test -tags=integration ./internal/notification/infrastructure/...
//
// Requires EDU_TEST_POSTGRES_URL and EDU_TEST_REDIS_URL; tests skip if unset.
package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	"github.com/son-ngo/edu-app/pkg/postgres"
	"github.com/son-ngo/edu-app/pkg/redis"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("EDU_TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("EDU_TEST_POSTGRES_URL not set")
	}
	pool, err := postgres.Connect(context.Background(), postgres.Config{URL: url})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedUser inserts a user row so device_token's FK is satisfied.
func seedUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.NewString()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, email, display_name, created_at, updated_at) VALUES ($1,$2,$3,NOW(),NOW())`,
		id, id+"@example.com", "Test")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func TestPgRepository_DeviceTokenLifecycle(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	userID := seedUser(t, pool)

	dt, _ := domain.NewDeviceToken(uuid.NewString(), userID, uuid.NewString(), domain.PlatformAndroid, time.Now())
	if err := repo.UpsertDeviceToken(ctx, dt); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	tok, err := repo.FindActiveDeviceToken(ctx, userID)
	if err != nil || tok != dt.Token {
		t.Fatalf("find active token: %v / %q", err, tok)
	}
	if err := repo.DeactivateToken(ctx, dt.Token); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if _, err := repo.FindActiveDeviceToken(ctx, userID); err == nil {
		t.Fatalf("expected no active token after deactivation")
	}
}

func TestPgRepository_PreferenceAndLog(t *testing.T) {
	pool := testPool(t)
	repo := NewPgRepository(pool)
	ctx := context.Background()
	userID := seedUser(t, pool)

	// Preference upsert + read.
	if err := repo.UpsertPreference(ctx, &domain.NotificationPreference{UserID: userID, Type: domain.TypeDailyReminder, Enabled: false}); err != nil {
		t.Fatalf("upsert pref: %v", err)
	}
	pref, err := repo.FindPreference(ctx, userID, domain.TypeDailyReminder)
	if err != nil || pref.Enabled {
		t.Fatalf("expected disabled pref, got %+v / %v", pref, err)
	}

	// Log save + status update + list.
	log := domain.NewPendingLog(uuid.NewString(), userID, "DAILY_REMINDER_V1", domain.TypeDailyReminder, uuid.NewString(), time.Now())
	if err := repo.SaveLog(ctx, log); err != nil {
		t.Fatalf("save log: %v", err)
	}
	sentAt := time.Now()
	if err := repo.UpdateLogStatus(ctx, log.ID, domain.StatusSent, &sentAt, ""); err != nil {
		t.Fatalf("update log: %v", err)
	}
	logs, total, err := repo.ListLogsByUser(ctx, userID, 10, 0)
	if err != nil || total < 1 || len(logs) < 1 {
		t.Fatalf("list logs: %v total=%d", err, total)
	}
	if logs[0].Status != domain.StatusSent {
		t.Fatalf("expected SENT, got %s", logs[0].Status)
	}
}

func TestRedisIdempotencyStore_DedupesWithinTTL(t *testing.T) {
	url := os.Getenv("EDU_TEST_REDIS_URL")
	if url == "" {
		t.Skip("EDU_TEST_REDIS_URL not set")
	}
	client, err := redis.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	store := NewRedisIdempotencyStore(client)
	key := uuid.NewString()

	first, err := store.CheckAndSet(context.Background(), key, time.Minute)
	if err != nil || !first {
		t.Fatalf("expected first-seen true, got %v / %v", first, err)
	}
	second, err := store.CheckAndSet(context.Background(), key, time.Minute)
	if err != nil || second {
		t.Fatalf("expected duplicate (false), got %v / %v", second, err)
	}
}
