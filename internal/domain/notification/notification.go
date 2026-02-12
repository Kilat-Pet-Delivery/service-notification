package notification

import (
	"time"

	"github.com/google/uuid"
)

// Notification is the aggregate root for a dispatched notification.
type Notification struct {
	id             uuid.UUID
	userID         uuid.UUID
	bookingID      *uuid.UUID
	eventType      string
	title          string
	body           string
	channelsSent   []NotificationChannel
	channelsFailed []NotificationChannel
	retryCount     int
	maxRetries     int
	status         NotificationStatus
	isRead         bool
	metadata       map[string]interface{}
	fcmSentAt      *time.Time
	smsSentAt      *time.Time
	emailSentAt    *time.Time
	version        int64
	createdAt      time.Time
	updatedAt      time.Time
}

// NewNotification creates a new notification aggregate.
func NewNotification(
	userID uuid.UUID,
	bookingID *uuid.UUID,
	eventType, title, body string,
	metadata map[string]interface{},
) *Notification {
	now := time.Now().UTC()
	return &Notification{
		id:             uuid.New(),
		userID:         userID,
		bookingID:      bookingID,
		eventType:      eventType,
		title:          title,
		body:           body,
		channelsSent:   []NotificationChannel{},
		channelsFailed: []NotificationChannel{},
		retryCount:     0,
		maxRetries:     3,
		status:         StatusPending,
		isRead:         false,
		metadata:       metadata,
		version:        1,
		createdAt:      now,
		updatedAt:      now,
	}
}

// Reconstitute rebuilds a notification from persistence.
func Reconstitute(
	id, userID uuid.UUID,
	bookingID *uuid.UUID,
	eventType, title, body string,
	channelsSent, channelsFailed []NotificationChannel,
	retryCount, maxRetries int,
	status NotificationStatus,
	isRead bool,
	metadata map[string]interface{},
	fcmSentAt, smsSentAt, emailSentAt *time.Time,
	version int64,
	createdAt, updatedAt time.Time,
) *Notification {
	return &Notification{
		id:             id,
		userID:         userID,
		bookingID:      bookingID,
		eventType:      eventType,
		title:          title,
		body:           body,
		channelsSent:   channelsSent,
		channelsFailed: channelsFailed,
		retryCount:     retryCount,
		maxRetries:     maxRetries,
		status:         status,
		isRead:         isRead,
		metadata:       metadata,
		fcmSentAt:      fcmSentAt,
		smsSentAt:      smsSentAt,
		emailSentAt:    emailSentAt,
		version:        version,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// MarkChannelSent records a successful send for the given channel.
func (n *Notification) MarkChannelSent(channel NotificationChannel) {
	n.channelsSent = append(n.channelsSent, channel)
	now := time.Now().UTC()
	switch channel {
	case ChannelPush:
		n.fcmSentAt = &now
	case ChannelSMS:
		n.smsSentAt = &now
	case ChannelEmail:
		n.emailSentAt = &now
	}
	n.refreshStatus()
	n.version++
	n.updatedAt = now
}

// MarkChannelFailed records a failed send for the given channel.
func (n *Notification) MarkChannelFailed(channel NotificationChannel) {
	n.channelsFailed = append(n.channelsFailed, channel)
	n.refreshStatus()
	n.version++
	n.updatedAt = time.Now().UTC()
}

// MarkAsRead marks the notification as read by the user.
func (n *Notification) MarkAsRead() {
	n.isRead = true
	n.version++
	n.updatedAt = time.Now().UTC()
}

// CanRetry returns true if the notification has failed channels and retries remaining.
func (n *Notification) CanRetry() bool {
	return n.retryCount < n.maxRetries && n.status != StatusSent
}

// IncrementRetry increments the retry counter.
func (n *Notification) IncrementRetry() {
	n.retryCount++
	n.version++
	n.updatedAt = time.Now().UTC()
}

func (n *Notification) refreshStatus() {
	sent := len(n.channelsSent)
	failed := len(n.channelsFailed)
	switch {
	case sent > 0 && failed == 0:
		n.status = StatusSent
	case sent > 0 && failed > 0:
		n.status = StatusPartiallySent
	case failed > 0 && sent == 0:
		n.status = StatusFailed
	}
}

// Getters.
func (n *Notification) ID() uuid.UUID                       { return n.id }
func (n *Notification) UserID() uuid.UUID                   { return n.userID }
func (n *Notification) BookingID() *uuid.UUID               { return n.bookingID }
func (n *Notification) EventType() string                   { return n.eventType }
func (n *Notification) Title() string                       { return n.title }
func (n *Notification) Body() string                        { return n.body }
func (n *Notification) ChannelsSent() []NotificationChannel { return n.channelsSent }
func (n *Notification) ChannelsFailed() []NotificationChannel {
	return n.channelsFailed
}
func (n *Notification) RetryCount() int            { return n.retryCount }
func (n *Notification) MaxRetries() int            { return n.maxRetries }
func (n *Notification) Status() NotificationStatus { return n.status }
func (n *Notification) IsRead() bool               { return n.isRead }
func (n *Notification) Metadata() map[string]interface{} {
	return n.metadata
}
func (n *Notification) FCMSentAt() *time.Time   { return n.fcmSentAt }
func (n *Notification) SMSSentAt() *time.Time   { return n.smsSentAt }
func (n *Notification) EmailSentAt() *time.Time { return n.emailSentAt }
func (n *Notification) Version() int64          { return n.version }
func (n *Notification) CreatedAt() time.Time    { return n.createdAt }
func (n *Notification) UpdatedAt() time.Time    { return n.updatedAt }
