package notification

// NotificationChannel represents a delivery channel for notifications.
type NotificationChannel string

const (
	ChannelPush  NotificationChannel = "push"
	ChannelSMS   NotificationChannel = "sms"
	ChannelEmail NotificationChannel = "email"
)

// AllChannels returns all available channels.
func AllChannels() []NotificationChannel {
	return []NotificationChannel{ChannelPush, ChannelSMS, ChannelEmail}
}
