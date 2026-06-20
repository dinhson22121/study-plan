package application

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	mu          sync.Mutex
	tokens      map[string]string
	templates   map[string]*domain.NotificationTemplate
	prefs       map[string]*domain.NotificationPreference
	logs        []*domain.NotificationLog
	activeUsers []string
	deactivated []string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		tokens:    map[string]string{},
		templates: map[string]*domain.NotificationTemplate{},
		prefs:     map[string]*domain.NotificationPreference{},
	}
}

func prefKey(userID string, t domain.NotificationType) string { return userID + "|" + string(t) }

func nowZero() time.Time { return time.Unix(0, 0).UTC() }

func (r *fakeRepo) UpsertDeviceToken(_ context.Context, dt *domain.DeviceToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[dt.UserID] = dt.Token
	return nil
}
func (r *fakeRepo) FindActiveDeviceToken(_ context.Context, userID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tok, ok := r.tokens[userID]; ok {
		return tok, nil
	}
	return "", shared.ErrNotFound
}
func (r *fakeRepo) DeactivateToken(_ context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deactivated = append(r.deactivated, token)
	return nil
}
func (r *fakeRepo) DeleteDeviceToken(_ context.Context, userID, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, userID)
	return nil
}
func (r *fakeRepo) FindTemplate(_ context.Context, code string) (*domain.NotificationTemplate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.templates[code]; ok {
		return t, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) FindPreference(_ context.Context, userID string, t domain.NotificationType) (*domain.NotificationPreference, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.prefs[prefKey(userID, t)]; ok {
		return p, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) ListPreferences(_ context.Context, userID string) ([]domain.NotificationPreference, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.NotificationPreference
	for _, p := range r.prefs {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (r *fakeRepo) UpsertPreference(_ context.Context, p *domain.NotificationPreference) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *p
	r.prefs[prefKey(p.UserID, p.Type)] = &cp
	return nil
}
func (r *fakeRepo) SaveLog(_ context.Context, l *domain.NotificationLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *l
	r.logs = append(r.logs, &cp)
	return nil
}
func (r *fakeRepo) UpdateLogStatus(_ context.Context, id string, status domain.NotificationStatus, sentAt *time.Time, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, l := range r.logs {
		if l.ID == id {
			l.Status = status
			l.SentAt = sentAt
			l.ErrorMessage = errMsg
			return nil
		}
	}
	return shared.ErrNotFound
}
func (r *fakeRepo) ListLogsByUser(_ context.Context, userID string, limit, offset int) ([]domain.NotificationLog, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var all []domain.NotificationLog
	for _, l := range r.logs {
		if l.UserID == userID {
			all = append(all, *l)
		}
	}
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}
func (r *fakeRepo) ListActiveUserIDs(_ context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.activeUsers...), nil
}

func (r *fakeRepo) logsByStatus(s domain.NotificationStatus) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, l := range r.logs {
		if l.Status == s {
			n++
		}
	}
	return n
}

type fakePublisher struct {
	mu       sync.Mutex
	messages []publishedMsg
	failNext error
}

type publishedMsg struct {
	Topic string
	Key   string
	Value []byte
}

func (p *fakePublisher) Publish(_ context.Context, topic string, key, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.failNext != nil {
		err := p.failNext
		p.failNext = nil
		return err
	}
	p.messages = append(p.messages, publishedMsg{Topic: topic, Key: string(key), Value: append([]byte(nil), value...)})
	return nil
}
func (p *fakePublisher) onTopic(topic string) []publishedMsg {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []publishedMsg
	for _, m := range p.messages {
		if m.Topic == topic {
			out = append(out, m)
		}
	}
	return out
}
func decode[T any](b []byte) T {
	var v T
	_ = json.Unmarshal(b, &v)
	return v
}

type fakeIdem struct {
	mu   sync.Mutex
	seen map[string]bool
}

func newFakeIdem() *fakeIdem { return &fakeIdem{seen: map[string]bool{}} }
func (i *fakeIdem) CheckAndSet(_ context.Context, key string, _ time.Duration) (bool, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.seen[key] {
		return false, nil
	}
	i.seen[key] = true
	return true, nil
}

type fakeFCM struct {
	err   error
	calls int
}

func (f *fakeFCM) Send(_ context.Context, _, _, _ string, _ map[string]string) error {
	f.calls++
	return f.err
}
