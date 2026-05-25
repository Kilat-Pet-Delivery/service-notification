package events

import (
	"context"
	"strings"

	"github.com/Kilat-Pet-Delivery/lib-common/kafka"
	protoEvents "github.com/Kilat-Pet-Delivery/lib-proto/events"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/categories"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// ChatEventConsumer listens to chat events and dispatches offline-recipient pushes.
type ChatEventConsumer struct {
	consumer *kafka.Consumer
	service  notificationDispatcher
	logger   *zap.Logger
}

func NewChatEventConsumer(brokers []string, groupID string, service notificationDispatcher, logger *zap.Logger) *ChatEventConsumer {
	return &ChatEventConsumer{
		consumer: kafka.NewConsumer(brokers, groupID, protoEvents.TopicChatEvents, logger),
		service:  service,
		logger:   logger,
	}
}

func (c *ChatEventConsumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *ChatEventConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	cloudEvent, err := kafka.ParseCloudEvent(msg.Value)
	if err != nil {
		return err
	}
	if !strings.EqualFold(cloudEvent.Type, protoEvents.ChatMessageSent) {
		c.logger.Debug("ignoring unhandled chat event", zap.String("type", cloudEvent.Type))
		return nil
	}
	var evt protoEvents.ChatMessageSentEvent
	if err := cloudEvent.ParseData(&evt); err != nil {
		return err
	}
	if evt.RecipientOnlineAtSendTime {
		return nil
	}
	return c.service.HandleEvent(ctx, categories.ChatMessage, evt.RecipientUserID, nil, map[string]interface{}{
		"ThreadID":  evt.ThreadID.String(),
		"MessageID": evt.MessageID.String(),
		"SenderID":  evt.SenderUserID.String(),
	})
}

func (c *ChatEventConsumer) Close() error {
	return c.consumer.Close()
}
