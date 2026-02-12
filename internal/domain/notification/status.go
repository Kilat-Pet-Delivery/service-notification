package notification

// NotificationStatus represents the delivery status of a notification.
type NotificationStatus string

const (
	StatusPending       NotificationStatus = "pending"
	StatusSent          NotificationStatus = "sent"
	StatusFailed        NotificationStatus = "failed"
	StatusPartiallySent NotificationStatus = "partially_sent"
)
