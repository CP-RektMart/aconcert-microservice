package domain

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/hub"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/repository"
	"github.com/google/uuid"
)

type Domain struct {
	hub  *hub.Hub
	repo repository.Repository
}

func New(hub *hub.Hub, repo repository.Repository) *Domain {
	return &Domain{
		hub:  hub,
		repo: repo,
	}
}

func (d *Domain) sendStreamData(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, eventType entities.EventType, data any) error {
	payload, err := json.Marshal(dto.EventStream{
		EventID: eventID,
		Data:    data,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal payload")
	}

	d.hub.Broadcast(ctx, userID, eventType, string(payload))

	return nil
}

func (d *Domain) PushMessage(ctx context.Context, userID uuid.UUID, eventType entities.EventType, data any) error {
	// 1. Create Event ID
	eventID := uuid.New()
	eventData := entities.EventData{
		EventID:   eventID,
		UserID:    userID,
		EventType: eventType,
		Data:      data,
	}

	// 2. Store Event Data
	if err := d.repo.SetEvent(ctx, eventData); err != nil {
		return errors.Wrap(err, "failed to store event")
	}

	// 3. Store User Events
	if err := d.repo.AddUserEvent(ctx, userID, eventID); err != nil {
		return errors.Wrap(err, "failed to store user event")
	}

	// 4. Broadcast the message to the user's connected clients
	d.sendStreamData(ctx, userID, eventID, eventType, data)

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

func (d *Domain) ReceiveAck(ctx context.Context, eventID uuid.UUID) error {
	// 1. Get Event ID
	event, err := d.repo.GetEvent(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "failed to get event")
	}

	// 2. Remove Event Cache
	if err := d.repo.RemoveEvent(ctx, eventID); err != nil {
		return errors.Wrap(err, "failed to remove event")
	}

	// 2. Remove User Event
	if err := d.repo.RemoveUserEvent(ctx, event.UserID, eventID); err != nil {
		return errors.Wrap(err, "failed to remove user event")
	}

	return nil
}
