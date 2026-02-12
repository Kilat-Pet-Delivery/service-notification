package adapter

import (
	"context"

	"go.uber.org/zap"
)

// TwilioAdapter defines the Anti-Corruption Layer interface for Twilio SMS.
type TwilioAdapter interface {
	SendSMS(ctx context.Context, to, message string) error
}

// MockTwilioAdapter is a development/testing implementation.
type MockTwilioAdapter struct {
	logger *zap.Logger
}

// NewMockTwilioAdapter creates a new mock Twilio adapter.
func NewMockTwilioAdapter(logger *zap.Logger) *MockTwilioAdapter {
	return &MockTwilioAdapter{logger: logger}
}

// SendSMS logs the SMS instead of sending it.
func (m *MockTwilioAdapter) SendSMS(ctx context.Context, to, message string) error {
	m.logger.Info("[MOCK TWILIO] SMS sent",
		zap.String("to", to),
		zap.String("message", message),
	)
	return nil
}
