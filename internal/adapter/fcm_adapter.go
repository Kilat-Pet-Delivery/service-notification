package adapter

import (
	"context"

	"go.uber.org/zap"
)

// FCMAdapter defines the Anti-Corruption Layer interface for Firebase Cloud Messaging.
type FCMAdapter interface {
	SendPush(ctx context.Context, token, title, body string, data map[string]string) error
}

// MockFCMAdapter is a development/testing implementation.
type MockFCMAdapter struct {
	logger *zap.Logger
}

// NewMockFCMAdapter creates a new mock FCM adapter.
func NewMockFCMAdapter(logger *zap.Logger) *MockFCMAdapter {
	return &MockFCMAdapter{logger: logger}
}

// SendPush logs the push notification instead of sending it.
func (m *MockFCMAdapter) SendPush(ctx context.Context, token, title, body string, data map[string]string) error {
	m.logger.Info("[MOCK FCM] Push notification sent",
		zap.String("token", token),
		zap.String("title", title),
		zap.String("body", body),
	)
	return nil
}
