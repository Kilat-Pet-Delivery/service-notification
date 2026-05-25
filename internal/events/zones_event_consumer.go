package events

import (
	"context"
	"strings"

	"github.com/Kilat-Pet-Delivery/lib-common/kafka"
	protoEvents "github.com/Kilat-Pet-Delivery/lib-proto/events"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/categories"
	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// ZoneRunnerProvider resolves online runner IDs for a zone.
type ZoneRunnerProvider interface {
	OnlineRunnerIDs(ctx context.Context, zoneID uuid.UUID) ([]uuid.UUID, error)
}

type noopZoneRunnerProvider struct{}

func (noopZoneRunnerProvider) OnlineRunnerIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

// ZonesEventConsumer listens to zone surge events and dispatches runner pushes.
type ZonesEventConsumer struct {
	consumer *kafka.Consumer
	service  notificationDispatcher
	runners  ZoneRunnerProvider
	logger   *zap.Logger
}

func NewZonesEventConsumer(brokers []string, groupID string, service notificationDispatcher, runners ZoneRunnerProvider, logger *zap.Logger) *ZonesEventConsumer {
	if runners == nil {
		runners = noopZoneRunnerProvider{}
	}
	return &ZonesEventConsumer{
		consumer: kafka.NewConsumer(brokers, groupID, protoEvents.TopicZonesEvents, logger),
		service:  service,
		runners:  runners,
		logger:   logger,
	}
}

func (c *ZonesEventConsumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *ZonesEventConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	cloudEvent, err := kafka.ParseCloudEvent(msg.Value)
	if err != nil {
		return err
	}
	if !strings.EqualFold(cloudEvent.Type, protoEvents.ZoneSurgeChanged) {
		return nil
	}
	var evt protoEvents.ZoneSurgeChangedEvent
	if err := cloudEvent.ParseData(&evt); err != nil {
		return err
	}
	if evt.NewMultiplier < 1.5 {
		return nil
	}
	runnerIDs, err := c.runners.OnlineRunnerIDs(ctx, evt.ZoneID)
	if err != nil {
		return err
	}
	for _, runnerID := range runnerIDs {
		if err := c.service.HandleEvent(ctx, categories.SurgeActive, runnerID, nil, map[string]interface{}{
			"ZoneID":     evt.ZoneID.String(),
			"ZoneCode":   evt.ZoneCode,
			"Multiplier": evt.NewMultiplier,
		}); err != nil {
			c.logger.Error("failed to dispatch surge notification", zap.String("runner_id", runnerID.String()), zap.Error(err))
		}
	}
	return nil
}

func (c *ZonesEventConsumer) Close() error {
	return c.consumer.Close()
}
