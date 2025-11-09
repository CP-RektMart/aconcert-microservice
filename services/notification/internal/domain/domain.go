package domain

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/pkg/realtime"
	"github.com/google/uuid"
)

type Domain struct {
	realtime *realtime.Service
}

func New(realtime *realtime.Service) *Domain {
	return &Domain{
		realtime: realtime,
	}
}

func (d *Domain) EventCreatedEvent(ctx context.Context, event entities.Event) error {
	if err := d.realtime.PushMessage(ctx, uuid.Nil, "event.created", event); err != nil {
		return errors.Wrap(err, "failed to push event.created message")
	}

	return nil
}

func (d *Domain) EventUpdatedEvent(ctx context.Context) error {
	panic("unimplemented")
}

func (d *Domain) ReservationConfirmedEvent(ctx context.Context) error {
	panic("unimplemented")
}

func (d *Domain) ReservationCancelledEvent(ctx context.Context) error {
	panic("unimplemented")
}
