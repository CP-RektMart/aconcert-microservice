package dto

import (
	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/entities"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/moonrhythm/validator"
)

type HttpResponse[T any] struct {
	Result T `json:"result" validate:"required"`
}

type HttpError struct {
	Error string `json:"error" validate:"required"`
}

type RealtimeRequest struct {
	UserID uuid.UUID `params:"userId"`
	State  string    `params:"state"`
}

func (r *RealtimeRequest) Parse(c *fiber.Ctx) error {
	if err := c.QueryParser(r); err != nil {
		return errors.Wrap(err, "failed to parse request")
	}

	if err := r.Validate(); err != nil {
		return errors.Wrap(err, "failed to validate request")
	}

	return nil
}

func (r *RealtimeRequest) Validate() error {
	v := validator.New()
	v.Must(r.UserID != uuid.Nil, "userId is required")

	return errors.WithStack(v.Error())
}

type PushMessage struct {
	UserID    uuid.UUID          `json:"userId"`
	EventType entities.EventType `json:"eventType"`
	Data      any                `json:"data"`
}

func (r *PushMessage) Parse(c *fiber.Ctx) error {
	if err := c.BodyParser(r); err != nil {
		return errors.Wrap(err, "failed to parse request")
	}

	if err := r.Validate(); err != nil {
		return errors.Wrap(err, "failed to validate request")
	}

	return nil
}

func (r *PushMessage) Validate() error {
	v := validator.New()
	v.Must(r.UserID != uuid.Nil, "userId is required")
	v.Must(r.EventType != "", "eventType is required")
	v.Must(r.Data != nil, "data is required")

	return errors.WithStack(v.Error())
}

type AckPushMessage struct {
	EventID uuid.UUID `json:"eventId"`
}

func (r *AckPushMessage) Parse(c *fiber.Ctx) error {
	if err := c.BodyParser(r); err != nil {
		return errors.Wrap(err, "failed to parse request")
	}

	if err := r.Validate(); err != nil {
		return errors.Wrap(err, "failed to validate request")
	}

	return nil
}

func (r *AckPushMessage) Validate() error {
	v := validator.New()
	v.Must(r.EventID != uuid.Nil, "eventId is required")

	return errors.WithStack(v.Error())
}

type EventStream struct {
	EventID uuid.UUID `json:"eventId"`
	Data    any       `json:"data"`
}
