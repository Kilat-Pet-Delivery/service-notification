package events

import (
	"context"

	"github.com/google/uuid"
)

type notificationDispatcher interface {
	HandleEvent(ctx context.Context, eventType string, userID uuid.UUID, bookingID *uuid.UUID, metadata map[string]interface{}) error
}
