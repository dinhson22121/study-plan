package infrastructure

import (
	"context"

	"go.uber.org/zap"
)

// LogSender is a pushSender fallback used when Firebase credentials are not
// configured. It logs what it would have sent and reports success, so the
// pipeline runs end-to-end in local development without real FCM. It is never
// used when a real FCM client initializes successfully.
type LogSender struct {
	log *zap.Logger
}

// NewLogSender builds the fallback sender.
func NewLogSender(log *zap.Logger) *LogSender { return &LogSender{log: log} }

// Send logs the notification and returns nil.
func (s *LogSender) Send(_ context.Context, token, title, body string, _ map[string]string) error {
	s.log.Info("LogSender: would send push (FCM not configured)",
		zap.String("token", token), zap.String("title", title), zap.String("body", body))
	return nil
}

// IsTokenInvalid always returns false: the fallback never invalidates tokens.
func (s *LogSender) IsTokenInvalid(error) bool { return false }
