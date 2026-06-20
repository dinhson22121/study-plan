package domain

import "time"

// NotificationLog is the aggregate tracking one notification's delivery
// lifecycle end-to-end, traceable by CorrelationID. State transitions return a
// new copy (immutable update) rather than mutating in place.
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

// NewPendingLog creates a log entry for a notification about to be sent.
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

// NewSkippedLog records that a notification was suppressed by user preference.
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

// MarkSent returns a copy transitioned to SENT at the given time.
func (l NotificationLog) MarkSent(at time.Time) *NotificationLog {
	l.Status = StatusSent
	l.SentAt = &at
	l.ErrorMessage = ""
	return &l
}

// MarkFailed returns a copy transitioned to FAILED with the error recorded.
func (l NotificationLog) MarkFailed(errMsg string) *NotificationLog {
	l.Status = StatusFailed
	l.ErrorMessage = errMsg
	return &l
}

// MarkRetrying returns a copy transitioned to RETRYING with an incremented count.
func (l NotificationLog) MarkRetrying() *NotificationLog {
	l.Status = StatusRetrying
	l.RetryCount++
	return &l
}
