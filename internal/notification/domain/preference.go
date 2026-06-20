package domain

type NotificationPreference struct {
	UserID  string
	Type    NotificationType
	Enabled bool
}

func DefaultPreferences(userID string) []NotificationPreference {
	types := AllTypes()
	prefs := make([]NotificationPreference, 0, len(types))
	for _, t := range types {
		prefs = append(prefs, NotificationPreference{UserID: userID, Type: t, Enabled: true})
	}
	return prefs
}
