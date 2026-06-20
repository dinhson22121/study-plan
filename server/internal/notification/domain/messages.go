package domain

const (
	TopicSchedule = "notification.schedule"
	TopicSend     = "notification.send"
	TopicResult   = "notification.result"
	TopicDLQ      = "notification.dlq"
)

func AllTopics() []string { return []string{TopicSchedule, TopicSend, TopicResult, TopicDLQ} }

const (
	ErrCodeTokenInvalid = "TOKEN_INVALID"
	ErrCodeMaxRetries   = "MAX_RETRIES"
	ErrCodeNoToken      = "NO_ACTIVE_TOKEN"
	ErrCodeTemplate     = "TEMPLATE_ERROR"
)

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

type SendMessage struct {
	CorrelationID string            `json:"correlationId"`
	UserID        string            `json:"userId"`
	DeviceToken   string            `json:"deviceToken"`
	Title         string            `json:"title"`
	Body          string            `json:"body"`
	Data          map[string]string `json:"data"`
	LogID         string            `json:"logId"`
}

type ResultMessage struct {
	CorrelationID string `json:"correlationId"`
	LogID         string `json:"logId"`
	Status        string `json:"status"`
	ErrorCode     string `json:"errorCode,omitempty"`
	ShouldRetry   bool   `json:"shouldRetry"`
	SentAt        string `json:"sentAt,omitempty"`
}
