package infrastructure

import (
	"context"

	"go.uber.org/zap"
)

type LogSender struct {
	log *zap.Logger
}

func NewLogSender(log *zap.Logger) *LogSender { return &LogSender{log: log} }

func (s *LogSender) Send(_ context.Context, token, title, body string, _ map[string]string) error {
	s.log.Info("LogSender: would send push (FCM not configured)",
		zap.String("token", token), zap.String("title", title), zap.String("body", body))
	return nil
}

func (s *LogSender) IsTokenInvalid(error) bool { return false }
