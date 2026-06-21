package observability

import (
	"time"

	"github.com/getsentry/sentry-go"
)

type SentryConfig struct {
	DSN         string
	Environment string
	Release     string
	SampleRate  float64
}

// InitSentry initialises the Sentry client when a DSN is configured. With an
// empty DSN it is a no-op: Sentry stays uninitialised, capture calls elsewhere
// (e.g. the Recovery middleware) silently drop, and the returned flush is safe
// to defer. The bool reports whether Sentry was actually enabled.
func InitSentry(cfg SentryConfig) (flush func(), enabled bool, err error) {
	if cfg.DSN == "" {
		return func() {}, false, nil
	}
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		TracesSampleRate: cfg.SampleRate,
		EnableTracing:    cfg.SampleRate > 0,
	})
	if err != nil {
		return func() {}, false, err
	}
	return func() { sentry.Flush(2 * time.Second) }, true, nil
}
