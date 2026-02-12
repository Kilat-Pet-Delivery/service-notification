package adapter

import (
	"context"

	"go.uber.org/zap"
)

// SMTPAdapter defines the Anti-Corruption Layer interface for email delivery.
type SMTPAdapter interface {
	SendEmail(ctx context.Context, to, subject, htmlBody string) error
}

// MockSMTPAdapter is a development/testing implementation.
type MockSMTPAdapter struct {
	logger *zap.Logger
}

// NewMockSMTPAdapter creates a new mock SMTP adapter.
func NewMockSMTPAdapter(logger *zap.Logger) *MockSMTPAdapter {
	return &MockSMTPAdapter{logger: logger}
}

// SendEmail logs the email instead of sending it.
func (m *MockSMTPAdapter) SendEmail(ctx context.Context, to, subject, htmlBody string) error {
	m.logger.Info("[MOCK SMTP] Email sent",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}
