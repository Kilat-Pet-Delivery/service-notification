package events

import (
	"context"

	"github.com/Kilat-Pet-Delivery/lib-common/kafka"
	protoEvents "github.com/Kilat-Pet-Delivery/lib-proto/events"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/application"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/categories"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// InventoryEventConsumer listens for inventory alerts.
type InventoryEventConsumer struct {
	consumer *kafka.Consumer
	service  *application.NotificationService
	logger   *zap.Logger
}

func NewInventoryEventConsumer(brokers []string, groupID string, service *application.NotificationService, logger *zap.Logger) *InventoryEventConsumer {
	return &InventoryEventConsumer{
		consumer: kafka.NewConsumer(brokers, groupID, protoEvents.TopicInventoryEvents, logger),
		service:  service,
		logger:   logger,
	}
}

func (c *InventoryEventConsumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *InventoryEventConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	ce, err := kafka.ParseCloudEvent(msg.Value)
	if err != nil {
		return err
	}
	if ce.Type != protoEvents.InventoryBelowThreshold {
		return nil
	}
	var evt protoEvents.InventoryBelowThresholdEvent
	if err := ce.ParseData(&evt); err != nil {
		return err
	}
	return c.service.HandleShopScopedEvent(ctx, categories.ShopLowStock, evt.ShopID, nil, map[string]interface{}{
		"ProductID":   evt.ProductID.String(),
		"ProductName": evt.Name,
		"StockOnHand": evt.StockOnHand,
		"Threshold":   evt.Threshold,
	})
}

func (c *InventoryEventConsumer) Close() error { return c.consumer.Close() }
