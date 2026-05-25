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

// LoyaltyEventConsumer listens to loyalty events and dispatches runner pushes.
type LoyaltyEventConsumer struct {
	consumer *kafka.Consumer
	service  notificationDispatcher
	logger   *zap.Logger
}

func NewLoyaltyEventConsumer(brokers []string, groupID string, service notificationDispatcher, logger *zap.Logger) *LoyaltyEventConsumer {
	return &LoyaltyEventConsumer{
		consumer: kafka.NewConsumer(brokers, groupID, protoEvents.TopicLoyaltyEvents, logger),
		service:  service,
		logger:   logger,
	}
}

func (c *LoyaltyEventConsumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *LoyaltyEventConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	cloudEvent, err := kafka.ParseCloudEvent(msg.Value)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(cloudEvent.Type, protoEvents.QuestCompleted):
		var evt protoEvents.QuestCompletedEvent
		if err := cloudEvent.ParseData(&evt); err != nil {
			return err
		}
		return c.service.HandleEvent(ctx, categories.QuestCompleted, evt.UserID, nil, map[string]interface{}{
			"QuestCode": evt.QuestCode,
			"Reward":    evt.RewardAmount,
			"Currency":  evt.Currency,
		})
	case strings.EqualFold(cloudEvent.Type, protoEvents.TierPromoted):
		var evt protoEvents.TierPromotedEvent
		if err := cloudEvent.ParseData(&evt); err != nil {
			return err
		}
		return c.service.HandleEvent(ctx, categories.TierPromoted, evt.UserID, nil, map[string]interface{}{
			"PreviousTier": evt.PreviousTier,
			"NewTier":      evt.NewTier,
		})
	default:
		c.logger.Debug("ignoring unhandled loyalty event", zap.String("type", cloudEvent.Type))
		return nil
	}
}

func (c *LoyaltyEventConsumer) Close() error {
	return c.consumer.Close()
}
