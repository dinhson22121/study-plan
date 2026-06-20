package domain

import "time"

type NotificationLog struct {
	ID            string
	UserID        string
	TemplateCode  string
	Type          NotificationType
	Status        NotificationStatus
	RetryCount    int
	CorrelationID string
	SentAt        *time.Time
	ErrorMessage  string
	CreatedAt     time.Time
}

func NewPendingLog(id, userID, templateCode string, t NotificationType, correlationID string, now time.Time) *NotificationLog {
	return &NotificationLog{
		ID:            id,
		UserID:        userID,
		TemplateCode:  templateCode,
		Type:          t,
		Status:        StatusPending,
		CorrelationID: correlationID,
		CreatedAt:     now,
	}
}

func NewSkippedLog(id, userID, templateCode string, t NotificationType, correlationID string, now time.Time) *NotificationLog {
	return &NotificationLog{
		ID:            id,
		UserID:        userID,
		TemplateCode:  templateCode,
		Type:          t,
		Status:        StatusSkipped,
		CorrelationID: correlationID,
		CreatedAt:     now,
	}
}

func (l NotificationLog) MarkSent(at time.Time) *NotificationLog {
	l.Status = StatusSent
	l.SentAt = &at
	l.ErrorMessage = ""
	return &l
}

func (l NotificationLog) MarkFailed(errMsg string) *NotificationLog {
	l.Status = StatusFailed
	l.ErrorMessage = errMsg
	return &l
}

func (l NotificationLog) MarkRetrying() *NotificationLog {
	l.Status = StatusRetrying
	l.RetryCount++
	return &l
}
