package domain

// Kafka topic names for the notification pipeline (PRD section 6).
const (
	TopicSchedule = "notification.schedule"
	TopicSend     = "notification.send"
	TopicResult   = "notification.result"
	TopicDLQ      = "notification.dlq"
)

// AllTopics lists every topic to provision at startup.
func AllTopics() []string { return []string{TopicSchedule, TopicSend, TopicResult, TopicDLQ} }

// Result error codes for notification.result messages.
const (
	ErrCodeTokenInvalid = "TOKEN_INVALID"
	ErrCodeMaxRetries   = "MAX_RETRIES"
	ErrCodeNoToken      = "NO_ACTIVE_TOKEN"
	ErrCodeTemplate     = "TEMPLATE_ERROR"
)

// ScheduleMessage is produced to notification.schedule by the scheduler and by
// upstream modules (study plan, quiz, progress) after the preference gate.
type ScheduleMessage struct {
	CorrelationID     string            `json:"correlationId"`
	StudentID         string            `json:"studentId"`
	NotificationType  NotificationType  `json:"notificationType"`
	TemplateCode      string            `json:"templateCode"`
	TemplateVariables map[string]string `json:"templateVariables"`
	ScheduledAt       string            `json:"scheduledAt"`
	IdempotencyKey    string            `json:"idempotencyKey"`
	DeepLink          string            `json:"deepLink,omitempty"`
}

// SendMessage is produced to notification.send after the token is resolved and
// the template is rendered.
type SendMessage struct {
	CorrelationID string            `json:"correlationId"`
	UserID        string            `json:"userId"`
	DeviceToken   string            `json:"deviceToken"`
	Title         string            `json:"title"`
	Body          string            `json:"body"`
	Data          map[string]string `json:"data"`
	LogID         string            `json:"logId"`
}

// ResultMessage is produced to notification.result by the FCM sender.
type ResultMessage struct {
	CorrelationID string `json:"correlationId"`
	LogID         string `json:"logId"`
	Status        string `json:"status"` // SENT | FAILED
	ErrorCode     string `json:"errorCode,omitempty"`
	ShouldRetry   bool   `json:"shouldRetry"`
	SentAt        string `json:"sentAt,omitempty"`
}
