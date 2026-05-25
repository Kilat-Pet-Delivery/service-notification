package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Kilat-Pet-Delivery/lib-common/kafka"
	protoEvents "github.com/Kilat-Pet-Delivery/lib-proto/events"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/categories"
	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type dispatchedNotification struct {
	eventType string
	userID    uuid.UUID
}

type fakeDispatcher struct {
	dispatched []dispatchedNotification
}

func (d *fakeDispatcher) HandleEvent(_ context.Context, eventType string, userID uuid.UUID, _ *uuid.UUID, _ map[string]interface{}) error {
	d.dispatched = append(d.dispatched, dispatchedNotification{eventType: eventType, userID: userID})
	return nil
}

type fakeZoneRunners struct {
	ids []uuid.UUID
}

func (r fakeZoneRunners) OnlineRunnerIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return r.ids, nil
}

func Test_QuestCompleted_PushesQuestCompletedCategory(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	consumer := &LoyaltyEventConsumer{service: dispatcher, logger: zap.NewNop()}
	userID := uuid.New()

	msg := eventMessage(t, protoEvents.QuestCompleted, protoEvents.QuestCompletedEvent{
		QuestID:    uuid.New(),
		QuestCode:  "five_runs",
		UserID:     userID,
		OccurredAt: time.Now().UTC(),
	})
	if err := consumer.handleMessage(context.Background(), msg); err != nil {
		t.Fatalf("handleMessage returned error: %v", err)
	}
	assertDispatched(t, dispatcher, categories.QuestCompleted, userID, 1)
}

func Test_SurgeChanged_PushesOnlyAtOrAbove1_5x(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	runnerID := uuid.New()
	consumer := &ZonesEventConsumer{service: dispatcher, runners: fakeZoneRunners{ids: []uuid.UUID{runnerID}}, logger: zap.NewNop()}

	low := eventMessage(t, protoEvents.ZoneSurgeChanged, protoEvents.ZoneSurgeChangedEvent{
		ZoneID:        uuid.New(),
		ZoneCode:      "klcc",
		NewMultiplier: 1.4,
		OccurredAt:    time.Now().UTC(),
	})
	if err := consumer.handleMessage(context.Background(), low); err != nil {
		t.Fatalf("low surge handleMessage returned error: %v", err)
	}
	if len(dispatcher.dispatched) != 0 {
		t.Fatalf("expected no push below 1.5x")
	}

	high := eventMessage(t, protoEvents.ZoneSurgeChanged, protoEvents.ZoneSurgeChangedEvent{
		ZoneID:        uuid.New(),
		ZoneCode:      "klcc",
		NewMultiplier: 1.5,
		OccurredAt:    time.Now().UTC(),
	})
	if err := consumer.handleMessage(context.Background(), high); err != nil {
		t.Fatalf("high surge handleMessage returned error: %v", err)
	}
	assertDispatched(t, dispatcher, categories.SurgeActive, runnerID, 1)
}

func Test_IncidentAssigned_PushesToAssignee(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	consumer := &IncidentEventConsumer{service: dispatcher, logger: zap.NewNop()}
	assigneeID := uuid.New()

	msg := eventMessage(t, protoEvents.IncidentAssigned, protoEvents.IncidentAssignedEvent{
		IncidentID:     uuid.New(),
		AssigneeUserID: assigneeID,
		OccurredAt:     time.Now().UTC(),
	})
	if err := consumer.handleMessage(context.Background(), msg); err != nil {
		t.Fatalf("handleMessage returned error: %v", err)
	}
	assertDispatched(t, dispatcher, categories.IncidentAssigned, assigneeID, 1)
}

func Test_ChatMessage_PushesOnly_WhenRecipientOnlineAtSendTimeIsFalse(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	consumer := &ChatEventConsumer{service: dispatcher, logger: zap.NewNop()}
	recipientID := uuid.New()

	online := chatMessageEvent(recipientID, true)
	if err := consumer.handleMessage(context.Background(), online); err != nil {
		t.Fatalf("online handleMessage returned error: %v", err)
	}
	if len(dispatcher.dispatched) != 0 {
		t.Fatalf("expected no push while recipient is online")
	}

	offline := chatMessageEvent(recipientID, false)
	if err := consumer.handleMessage(context.Background(), offline); err != nil {
		t.Fatalf("offline handleMessage returned error: %v", err)
	}
	assertDispatched(t, dispatcher, categories.ChatMessage, recipientID, 1)
}

func chatMessageEvent(recipientID uuid.UUID, online bool) kafkago.Message {
	return mustEventMessage(protoEvents.ChatMessageSent, protoEvents.ChatMessageSentEvent{
		MessageID:                 uuid.New(),
		ThreadID:                  uuid.New(),
		SenderUserID:              uuid.New(),
		RecipientUserID:           recipientID,
		RecipientOnlineAtSendTime: online,
		OccurredAt:                time.Now().UTC(),
	})
}

func eventMessage(t *testing.T, eventType string, data interface{}) kafkago.Message {
	t.Helper()
	return mustEventMessage(eventType, data)
}

func mustEventMessage(eventType string, data interface{}) kafkago.Message {
	cloudEvent, err := kafka.NewCloudEvent("test", eventType, data)
	if err != nil {
		panic(err)
	}
	raw, err := json.Marshal(cloudEvent)
	if err != nil {
		panic(err)
	}
	return kafkago.Message{Value: raw}
}

func assertDispatched(t *testing.T, dispatcher *fakeDispatcher, eventType string, userID uuid.UUID, count int) {
	t.Helper()
	if len(dispatcher.dispatched) != count {
		t.Fatalf("expected %d dispatched notifications, got %d", count, len(dispatcher.dispatched))
	}
	last := dispatcher.dispatched[len(dispatcher.dispatched)-1]
	if last.eventType != eventType || last.userID != userID {
		t.Fatalf("expected %s for %s, got %+v", eventType, userID, last)
	}
}
