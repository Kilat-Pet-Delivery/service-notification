package notification

import (
	"context"

	"github.com/google/uuid"
)

// NotificationRepository defines persistence operations for notifications.
type NotificationRepository interface {
	Save(ctx context.Context, n *Notification) error
	Update(ctx context.Context, n *Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	FindPendingRetries(ctx context.Context) ([]*Notification, error)
	CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// PreferenceRepository defines persistence operations for notification preferences.
type PreferenceRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*NotificationPreference, error)
	Save(ctx context.Context, pref *NotificationPreference) error
	Update(ctx context.Context, pref *NotificationPreference) error
	Upsert(ctx context.Context, pref *NotificationPreference) error
}
