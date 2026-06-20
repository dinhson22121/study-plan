package application

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
)

func newManager(repo *fakeRepo, pub *fakePublisher) *Manager {
	d := NewDispatcher(repo, newFakeIdem(), pub, zap.NewNop())
	return NewManager(repo, d)
}

func TestManager_ListPreferencesDefaultsWhenEmpty(t *testing.T) {
	repo := newFakeRepo()
	m := newManager(repo, &fakePublisher{})

	prefs, err := m.ListPreferences(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prefs) != len(domain.AllTypes()) {
		t.Fatalf("expected default prefs for all types, got %d", len(prefs))
	}
}

func TestManager_SetPreferenceValidatesType(t *testing.T) {
	m := newManager(newFakeRepo(), &fakePublisher{})
	if err := m.SetPreference(context.Background(), "u1", "BOGUS", false); err == nil {
		t.Fatalf("expected error for invalid type")
	}
	if err := m.SetPreference(context.Background(), "u1", string(domain.TypeDailyReminder), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManager_RegisterDeviceTokenValidates(t *testing.T) {
	m := newManager(newFakeRepo(), &fakePublisher{})
	if err := m.RegisterDeviceToken(context.Background(), "u1", "tok", "windows"); err == nil {
		t.Fatalf("expected error for bad platform")
	}
	if err := m.RegisterDeviceToken(context.Background(), "u1", "tok", "android"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManager_BroadcastFansOutThroughGate(t *testing.T) {
	repo := newFakeRepo()
	repo.activeUsers = []string{"u1", "u2", "u3"}
	// u2 disabled this type — should be skipped by the dispatcher gate.
	_ = repo.UpsertPreference(context.Background(), &domain.NotificationPreference{
		UserID: "u2", Type: domain.TypeAdminBroadcast, Enabled: false,
	})
	pub := &fakePublisher{}
	m := newManager(repo, pub)

	count, err := m.Broadcast(context.Background(), BroadcastInput{
		Type: domain.TypeAdminBroadcast, TemplateCode: "ADMIN_V1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 users processed, got %d", count)
	}
	// Only u1 and u3 actually enqueue (u2 gated off).
	if got := len(pub.onTopic(domain.TopicSchedule)); got != 2 {
		t.Fatalf("expected 2 enqueued (u2 gated), got %d", got)
	}
}

func TestManager_GetHistoryPaginates(t *testing.T) {
	repo := newFakeRepo()
	for i := 0; i < 5; i++ {
		_ = repo.SaveLog(context.Background(), domain.NewPendingLog("l"+string(rune('0'+i)), "u1", "T", domain.TypeDailyReminder, "c", nowZero()))
	}
	m := newManager(repo, &fakePublisher{})

	page, total, err := m.GetHistory(context.Background(), "u1", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 || len(page) != 2 {
		t.Fatalf("expected total 5 and page size 2, got total=%d page=%d", total, len(page))
	}
}
