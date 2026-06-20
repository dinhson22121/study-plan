package domain

// NotificationPreference is a per-user, per-type on/off switch. The scheduler
// consults it before enqueuing any notification (the "preference gate").
type NotificationPreference struct {
	UserID  string
	Type    NotificationType
	Enabled bool
}

// DefaultPreferences returns all types enabled — the default for a new user.
func DefaultPreferences(userID string) []NotificationPreference {
	types := AllTypes()
	prefs := make([]NotificationPreference, 0, len(types))
	for _, t := range types {
		prefs = append(prefs, NotificationPreference{UserID: userID, Type: t, Enabled: true})
	}
	return prefs
}
