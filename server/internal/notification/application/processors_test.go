package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func seedTemplate(repo *fakeRepo, code string) {
	repo.templates[code] = &domain.NotificationTemplate{
		Code: code, Title: "Hi {name}", Body: "Streak {streak}", Type: domain.TypeDailyReminder, IsActive: true,
	}
}

func TestScheduleProcessor_ResolvesTokenRendersAndProducesSend(t *testing.T) {
	repo, pub := newFakeRepo(), &fakePublisher{}
	_ = repo.UpsertDeviceToken(context.Background(), &domain.DeviceToken{UserID: "u1", Token: "tok", IsActive: true})
	seedTemplate(repo, "T")
	p := NewScheduleProcessor(repo, pub, zap.NewNop())

	err := p.Process(context.Background(), domain.ScheduleMessage{
		CorrelationID: "c1", StudentID: "u1", NotificationType: domain.TypeDailyReminder,
		TemplateCode: "T", TemplateVariables: map[string]string{"name": "Minh", "streak": "7"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sends := pub.onTopic(domain.TopicSend)
	if len(sends) != 1 {
		t.Fatalf("expected 1 send message, got %d", len(sends))
	}
	sm := decode[domain.SendMessage](sends[0].Value)
	if sm.Title != "Hi Minh" || sm.Body != "Streak 7" || sm.DeviceToken != "tok" {
		t.Fatalf("send message rendered wrong: %+v", sm)
	}
	if repo.logsByStatus(domain.StatusPending) != 1 {
		t.Fatalf("expected a PENDING log")
	}
}

func TestScheduleProcessor_NoTokenRecordsFailedNoSend(t *testing.T) {
	repo, pub := newFakeRepo(), &fakePublisher{}
	seedTemplate(repo, "T")
	p := NewScheduleProcessor(repo, pub, zap.NewNop())

	err := p.Process(context.Background(), domain.ScheduleMessage{StudentID: "ghost", TemplateCode: "T", NotificationType: domain.TypeDailyReminder})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pub.onTopic(domain.TopicSend)) != 0 {
		t.Fatalf("no token should not produce a send")
	}
	if repo.logsByStatus(domain.StatusFailed) != 1 {
		t.Fatalf("expected a FAILED log for missing token")
	}
}

func TestScheduleProcessor_MissingTemplateRecordsFailed(t *testing.T) {
	repo, pub := newFakeRepo(), &fakePublisher{}
	_ = repo.UpsertDeviceToken(context.Background(), &domain.DeviceToken{UserID: "u1", Token: "tok"})
	p := NewScheduleProcessor(repo, pub, zap.NewNop())

	_ = p.Process(context.Background(), domain.ScheduleMessage{StudentID: "u1", TemplateCode: "MISSING", NotificationType: domain.TypeDailyReminder})
	if repo.logsByStatus(domain.StatusFailed) != 1 {
		t.Fatalf("expected FAILED log for missing template")
	}
}

func TestSendProcessor_SuccessProducesSentResult(t *testing.T) {
	pub := &fakePublisher{}
	p := NewSendProcessor(&fakeFCM{err: nil}, pub, zap.NewNop())

	_ = p.Process(context.Background(), domain.SendMessage{LogID: "l1", DeviceToken: "tok", Title: "t", Body: "b"})
	results := pub.onTopic(domain.TopicResult)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	rm := decode[domain.ResultMessage](results[0].Value)
	if rm.Status != string(domain.StatusSent) || rm.SentAt == "" {
		t.Fatalf("expected SENT result with timestamp, got %+v", rm)
	}
}

func TestSendProcessor_TokenInvalidProducesFailedResult(t *testing.T) {
	pub := &fakePublisher{}
	p := NewSendProcessor(&fakeFCM{err: shared.ErrTokenInvalid}, pub, zap.NewNop())

	_ = p.Process(context.Background(), domain.SendMessage{LogID: "l1"})
	rm := decode[domain.ResultMessage](pub.onTopic(domain.TopicResult)[0].Value)
	if rm.Status != string(domain.StatusFailed) || rm.ErrorCode != domain.ErrCodeTokenInvalid {
		t.Fatalf("expected FAILED/TOKEN_INVALID, got %+v", rm)
	}
}

func TestResultProcessor_SentUpdatesLog(t *testing.T) {
	repo, pub := newFakeRepo(), &fakePublisher{}
	_ = repo.SaveLog(context.Background(), domain.NewPendingLog("l1", "u1", "T", domain.TypeDailyReminder, "c", time.Unix(0, 0)))
	p := NewResultProcessor(repo, pub, zap.NewNop())

	err := p.Process(context.Background(), domain.ResultMessage{LogID: "l1", Status: string(domain.StatusSent), SentAt: time.Unix(2000, 0).UTC().Format(time.RFC3339)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.logsByStatus(domain.StatusSent) != 1 {
		t.Fatalf("expected log marked SENT")
	}
	if len(pub.onTopic(domain.TopicDLQ)) != 0 {
		t.Fatalf("SENT result must not dead-letter")
	}
}

func TestResultProcessor_FailedUpdatesLogAndDeadLetters(t *testing.T) {
	repo, pub := newFakeRepo(), &fakePublisher{}
	_ = repo.SaveLog(context.Background(), domain.NewPendingLog("l1", "u1", "T", domain.TypeDailyReminder, "c", time.Unix(0, 0)))
	p := NewResultProcessor(repo, pub, zap.NewNop())

	err := p.Process(context.Background(), domain.ResultMessage{LogID: "l1", Status: string(domain.StatusFailed), ErrorCode: domain.ErrCodeTokenInvalid})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.logsByStatus(domain.StatusFailed) != 1 {
		t.Fatalf("expected log marked FAILED")
	}
	if len(pub.onTopic(domain.TopicDLQ)) != 1 {
		t.Fatalf("FAILED result must dead-letter")
	}
}
