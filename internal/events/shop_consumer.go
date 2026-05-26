package events

import (
	"context"
	"strings"

	"github.com/Kilat-Pet-Delivery/lib-common/kafka"
	protoEvents "github.com/Kilat-Pet-Delivery/lib-proto/events"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/application"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// ShopEventConsumer listens for shop staff and merchant lifecycle events.
type ShopEventConsumer struct {
	consumer *kafka.Consumer
	service  *application.NotificationService
	logger   *zap.Logger
}

func NewShopEventConsumer(brokers []string, groupID string, service *application.NotificationService, logger *zap.Logger) *ShopEventConsumer {
	return &ShopEventConsumer{
		consumer: kafka.NewConsumer(brokers, groupID, protoEvents.TopicShopEvents, logger),
		service:  service,
		logger:   logger,
	}
}

func (c *ShopEventConsumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *ShopEventConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	cloudEvent, err := kafka.ParseCloudEvent(msg.Value)
	if err != nil {
		c.logger.Error("failed to parse cloud event from shop topic", zap.Error(err))
		return err
	}

	if strings.EqualFold(cloudEvent.Type, protoEvents.ShopStaffInvited) {
		return c.handleStaffInvited(ctx, cloudEvent)
	}
	return nil
}

func (c *ShopEventConsumer) handleStaffInvited(ctx context.Context, ce kafka.CloudEvent) error {
	var evt protoEvents.ShopStaffInvitedEvent
	if err := ce.ParseData(&evt); err != nil {
		return err
	}
	if evt.Email == "" {
		c.logger.Info("shop staff invite skipped (no email)", zap.String("invite_id", evt.InviteID.String()))
		return nil
	}

	return c.service.SendStaffInviteEmail(ctx, evt.Email, map[string]interface{}{
		"InviteID":  evt.InviteID.String(),
		"ShopID":    evt.ShopID.String(),
		"ShopName":  "your shop",
		"Role":      evt.Role,
		"Token":     evt.Token,
		"AcceptURL": "kilatpet://staff-invites/" + evt.Token,
	})
}

func (c *ShopEventConsumer) Close() error {
	return c.consumer.Close()
}
