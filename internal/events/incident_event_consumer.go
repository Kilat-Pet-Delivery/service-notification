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

// IncidentEventConsumer listens to incident events and dispatches support pushes.
type IncidentEventConsumer struct {
	consumer *kafka.Consumer
	service  notificationDispatcher
	logger   *zap.Logger
}

func NewIncidentEventConsumer(brokers []string, groupID string, service notificationDispatcher, logger *zap.Logger) *IncidentEventConsumer {
	return &IncidentEventConsumer{
		consumer: kafka.NewConsumer(brokers, groupID, protoEvents.TopicIncidentEvents, logger),
		service:  service,
		logger:   logger,
	}
}

func (c *IncidentEventConsumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *IncidentEventConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	cloudEvent, err := kafka.ParseCloudEvent(msg.Value)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(cloudEvent.Type, protoEvents.IncidentAssigned):
		var evt protoEvents.IncidentAssignedEvent
		if err := cloudEvent.ParseData(&evt); err != nil {
			return err
		}
		return c.service.HandleEvent(ctx, categories.IncidentAssigned, evt.AssigneeUserID, nil, map[string]interface{}{
			"IncidentID": evt.IncidentID.String(),
		})
	case strings.EqualFold(cloudEvent.Type, protoEvents.IncidentCreated):
		var evt protoEvents.IncidentCreatedEvent
		if err := cloudEvent.ParseData(&evt); err != nil {
			return err
		}
		if !strings.EqualFold(evt.Type, "sos") || !strings.EqualFold(evt.Severity, "critical") || evt.AssigneeUserID == nil {
			return nil
		}
		return c.service.HandleEvent(ctx, categories.SOSAck, *evt.AssigneeUserID, evt.BookingID, map[string]interface{}{
			"IncidentID": evt.IncidentID.String(),
		})
	default:
		c.logger.Debug("ignoring unhandled incident event", zap.String("type", cloudEvent.Type))
		return nil
	}
}

func (c *IncidentEventConsumer) Close() error {
	return c.consumer.Close()
}
