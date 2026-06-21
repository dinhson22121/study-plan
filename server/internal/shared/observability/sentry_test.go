package observability

import "testing"

func TestInitSentry_DisabledWhenNoDSN(t *testing.T) {
	flush, enabled, err := InitSentry(SentryConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enabled {
		t.Fatalf("expected Sentry to be disabled with an empty DSN")
	}
	if flush == nil {
		t.Fatalf("flush must be non-nil even when disabled")
	}
	flush() // must not panic
}

func TestInitSentry_EnabledWithDSN(t *testing.T) {
	flush, enabled, err := InitSentry(SentryConfig{
		DSN:         "https://public@example.ingest.sentry.io/1",
		Environment: "test",
		SampleRate:  0.1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !enabled {
		t.Fatalf("expected Sentry to be enabled with a valid DSN")
	}
	flush()
}
