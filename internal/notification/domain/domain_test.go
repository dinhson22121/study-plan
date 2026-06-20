package domain

import (
	"errors"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestTemplate_RenderSubstitutesVariables(t *testing.T) {
	tmpl := NotificationTemplate{
		Title: "Đã đến giờ học, {name}!",
		Body:  "Streak {streak} ngày của bạn đang chờ.",
	}
	title, body, err := tmpl.Render(map[string]string{"name": "Minh", "streak": "7"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "Đã đến giờ học, Minh!" {
		t.Fatalf("title not rendered: %q", title)
	}
	if body != "Streak 7 ngày của bạn đang chờ." {
		t.Fatalf("body not rendered: %q", body)
	}
}

func TestTemplate_RenderFailsOnMissingVariable(t *testing.T) {
	tmpl := NotificationTemplate{Title: "Hi {name}", Body: "x"}
	_, _, err := tmpl.Render(map[string]string{})
	if !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for missing var, got %v", err)
	}
}

func TestNotificationLog_Transitions(t *testing.T) {
	at := time.Unix(1000, 0)
	pending := NewPendingLog("l1", "u1", "TPL", TypeDailyReminder, "corr", at)
	if pending.Status != StatusPending {
		t.Fatalf("expected PENDING, got %s", pending.Status)
	}

	sent := pending.MarkSent(at)
	if sent.Status != StatusSent || sent.SentAt == nil {
		t.Fatalf("MarkSent failed: %+v", sent)
	}
	if pending.Status != StatusPending {
		t.Fatalf("MarkSent mutated the original (should be immutable)")
	}

	failed := pending.MarkFailed("boom")
	if failed.Status != StatusFailed || failed.ErrorMessage != "boom" {
		t.Fatalf("MarkFailed failed: %+v", failed)
	}

	retrying := pending.MarkRetrying()
	if retrying.Status != StatusRetrying || retrying.RetryCount != 1 {
		t.Fatalf("MarkRetrying failed: %+v", retrying)
	}
}

func TestParseType(t *testing.T) {
	if _, err := ParseType("DAILY_REMINDER"); err != nil {
		t.Fatalf("expected valid type, got %v", err)
	}
	if _, err := ParseType("NONSENSE"); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for unknown type")
	}
}

func TestDefaultPreferences_AllEnabled(t *testing.T) {
	prefs := DefaultPreferences("u1")
	if len(prefs) != len(AllTypes()) {
		t.Fatalf("expected one pref per type, got %d", len(prefs))
	}
	for _, p := range prefs {
		if !p.Enabled || p.UserID != "u1" {
			t.Fatalf("default pref should be enabled for u1: %+v", p)
		}
	}
}

func TestNewDeviceToken_Validation(t *testing.T) {
	now := time.Unix(0, 0)
	if _, err := NewDeviceToken("id", "", "tok", PlatformAndroid, now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for empty user id")
	}
	if _, err := NewDeviceToken("id", "u1", "tok", Platform("symbian"), now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for bad platform")
	}
	dt, err := NewDeviceToken("id", "u1", "tok", PlatformIOS, now)
	if err != nil || !dt.IsActive {
		t.Fatalf("expected valid active token, got %v / %+v", err, dt)
	}
}
