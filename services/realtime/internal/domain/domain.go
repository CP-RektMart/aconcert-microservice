package domain

import (
	"context"
	"encoding/json"

	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/hub"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/pubsub"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/repository"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Domain struct {
	hub             *hub.Hub
	repo            repository.Repository
	eventSubscriber *pubsub.EventSubscriber
}

func New(hub *hub.Hub, repo repository.Repository, eventSubscriber *pubsub.EventSubscriber) *Domain {
	return &Domain{
		hub:             hub,
		repo:            repo,
		eventSubscriber: eventSubscriber,
	}
}

// SubscribeToEvent registers a user's interest in event seat updates
func (d *Domain) SubscribeToEvent(ctx context.Context, userID uuid.UUID, eventID string) error {
	return d.eventSubscriber.Subscribe(ctx, userID, eventID)
}

// UnsubscribeFromEvent removes a user from event seat updates
func (d *Domain) UnsubscribeFromEvent(ctx context.Context, userID uuid.UUID, eventID string) error {
	return d.eventSubscriber.Unsubscribe(ctx, userID, eventID)
}

// UnsubscribeUserFromAll removes a user from all event subscriptions
func (d *Domain) UnsubscribeUserFromAll(ctx context.Context, userID uuid.UUID) error {
	return d.eventSubscriber.UnsubscribeUserFromAll(ctx, userID)
}

func (d *Domain) sendStreamData(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, eventType string, data any) error {
	payload, err := json.Marshal(dto.EventStream{
		EventID:   eventID,
		EventType: eventType,
		Data:      data,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal payload")
	}

	d.hub.Broadcast(ctx, userID, eventType, string(payload))

	return nil
}

func (d *Domain) broadcastStreamData(ctx context.Context, eventID uuid.UUID, eventType string, data any) error {
	payload, err := json.Marshal(dto.EventStream{
		EventID:   eventID,
		EventType: eventType,
		Data:      data,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal payload")
	}

	d.hub.BroadcastAll(ctx, eventType, string(payload))

	return nil
}

func (d *Domain) PushMessage(ctx context.Context, userID uuid.UUID, eventType string, data any) error {
	// 1. Create Event ID
	eventID := uuid.New()
	eventData := entities.EventData{
		EventID:   eventID,
		UserID:    userID,
		EventType: eventType,
		Data:      data,
	}

	// 2. Store Event Data
	// TODO: Remove?
	if err := d.repo.SetEvent(ctx, eventData); err != nil {
		return errors.Wrap(err, "failed to store event")
	}

	// 3. Store User Events
	// TODO: Remove?
	if err := d.repo.AddUserEvent(ctx, userID, eventID); err != nil {
		return errors.Wrap(err, "failed to store user event")
	}

	// 4. Broadcast the message to the user's connected clients
	if userID == uuid.Nil {
		if err := d.broadcastStreamData(ctx, eventID, eventType, data); err != nil {
			return errors.Wrap(err, "failed to broadcast to all clients")
		}
	} else {
		if err := d.sendStreamData(ctx, userID, eventID, eventType, data); err != nil {
			return errors.Wrap(err, "failed to send to user clients")
		}
	}

	return nil
}

func (d *Domain) ResendCache(ctx context.Context, userID uuid.UUID) error {
	events, err := d.repo.GetUserEvents(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user events")
	}

	for _, event := range events {
		d.sendStreamData(ctx, userID, event.EventID, event.EventType, event.Data)
	}

	return nil
}
