package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func publishJSON(ctx context.Context, pub domain.Publisher, topic, key string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if err := pub.Publish(ctx, topic, []byte(key), b); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

type ScheduleProcessor struct {
	repo  domain.Repository
	pub   domain.Publisher
	log   *zap.Logger
	now   func() time.Time
	newID func() string
}

func NewScheduleProcessor(repo domain.Repository, pub domain.Publisher, log *zap.Logger) *ScheduleProcessor {
	return &ScheduleProcessor{repo: repo, pub: pub, log: log, now: time.Now, newID: uuid.NewString}
}

func (p *ScheduleProcessor) Process(ctx context.Context, msg domain.ScheduleMessage) error {
	logID := p.newID()

	token, err := p.repo.FindActiveDeviceToken(ctx, msg.StudentID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return p.recordFailed(ctx, logID, msg, "no active device token")
		}
		return err
	}

	tmpl, err := p.repo.FindTemplate(ctx, msg.TemplateCode)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return p.recordFailed(ctx, logID, msg, "template not found: "+msg.TemplateCode)
		}
		return err
	}

	title, body, err := tmpl.Render(msg.TemplateVariables)
	if err != nil {
		return p.recordFailed(ctx, logID, msg, "template render failed: "+err.Error())
	}

	pending := domain.NewPendingLog(logID, msg.StudentID, msg.TemplateCode, msg.NotificationType, msg.CorrelationID, p.now())
	if err := p.repo.SaveLog(ctx, pending); err != nil {
		return err
	}

	data := map[string]string{}
	if msg.DeepLink != "" {
		data["deepLink"] = msg.DeepLink
	}
	send := domain.SendMessage{
		CorrelationID: msg.CorrelationID,
		UserID:        msg.StudentID,
		DeviceToken:   token,
		Title:         title,
		Body:          body,
		Data:          data,
		LogID:         logID,
	}
	return publishJSON(ctx, p.pub, domain.TopicSend, msg.StudentID, send)
}

func (p *ScheduleProcessor) recordFailed(ctx context.Context, logID string, msg domain.ScheduleMessage, reason string) error {
	failed := domain.NewPendingLog(logID, msg.StudentID, msg.TemplateCode, msg.NotificationType, msg.CorrelationID, p.now()).MarkFailed(reason)
	p.log.Warn("notification schedule failed", zap.String("user_id", msg.StudentID), zap.String("reason", reason))
	return p.repo.SaveLog(ctx, failed)
}

type SendProcessor struct {
	fcm domain.FCMPort
	pub domain.Publisher
	log *zap.Logger
	now func() time.Time
}

func NewSendProcessor(fcm domain.FCMPort, pub domain.Publisher, log *zap.Logger) *SendProcessor {
	return &SendProcessor{fcm: fcm, pub: pub, log: log, now: time.Now}
}

func (p *SendProcessor) Process(ctx context.Context, msg domain.SendMessage) error {
	result := domain.ResultMessage{CorrelationID: msg.CorrelationID, LogID: msg.LogID}

	err := p.fcm.Send(ctx, msg.DeviceToken, msg.Title, msg.Body, msg.Data)
	switch {
	case err == nil:
		result.Status = string(domain.StatusSent)
		result.SentAt = p.now().Format(time.RFC3339)
	case errors.Is(err, shared.ErrTokenInvalid):
		result.Status = string(domain.StatusFailed)
		result.ErrorCode = domain.ErrCodeTokenInvalid
	case errors.Is(err, shared.ErrMaxRetriesExceeded):
		result.Status = string(domain.StatusFailed)
		result.ErrorCode = domain.ErrCodeMaxRetries
	default:
		result.Status = string(domain.StatusFailed)
		result.ErrorCode = "SEND_ERROR"
	}

	return publishJSON(ctx, p.pub, domain.TopicResult, msg.LogID, result)
}

type ResultProcessor struct {
	repo domain.Repository
	pub  domain.Publisher
	log  *zap.Logger
}

func NewResultProcessor(repo domain.Repository, pub domain.Publisher, log *zap.Logger) *ResultProcessor {
	return &ResultProcessor{repo: repo, pub: pub, log: log}
}

func (p *ResultProcessor) Process(ctx context.Context, msg domain.ResultMessage) error {
	if msg.Status == string(domain.StatusSent) {
		var sentAt *time.Time
		if t, err := time.Parse(time.RFC3339, msg.SentAt); err == nil {
			sentAt = &t
		}
		return p.repo.UpdateLogStatus(ctx, msg.LogID, domain.StatusSent, sentAt, "")
	}

	if err := p.repo.UpdateLogStatus(ctx, msg.LogID, domain.StatusFailed, nil, msg.ErrorCode); err != nil {
		return err
	}
	p.log.Warn("notification dead-lettered", zap.String("log_id", msg.LogID), zap.String("error_code", msg.ErrorCode))
	return publishJSON(ctx, p.pub, domain.TopicDLQ, msg.LogID, msg)
}
