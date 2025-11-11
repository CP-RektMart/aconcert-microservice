package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cp-rektmart/aconcert-microservice/notification/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/entities"
	"github.com/streadway/amqp"
)

type Handler struct {
	domain *domain.Domain
}

func New(domain *domain.Domain) *Handler {
	return &Handler{
		domain: domain,
	}
}

func (h *Handler) Mount(ctx context.Context, msgs <-chan amqp.Delivery) error {
	for d := range msgs {
		var eventData entities.Message
		if err := json.Unmarshal(d.Body, &eventData); err != nil {
			log.Printf("Error reading event data (invalid JSON): %s", err)
			continue
		}

		fmt.Printf("📩 Received message: %+v\n", eventData.Type)
		switch eventData.Type {
		case entities.MessageTypeEventCreated:
			var event entities.Event
			if err := json.Unmarshal(eventData.Data, &event); err != nil {
				log.Printf("Error unmarshaling event.created payload: %s", err)
				continue
			}

			if err := h.domain.EventCreatedEvent(ctx, event); err != nil {
				log.Printf("Error handling event.created: %s", err)
			}
		case entities.MessageTypeReservationConfirmed:
			var reservation entities.ConfirmedNotiReservation
			if err := json.Unmarshal(eventData.Data, &reservation); err != nil {
				log.Printf("Error unmarshaling reservation.confirmed payload: %s", err)
				continue
			}

			if err := h.domain.ReservationConfirmedEvent(ctx, reservation); err != nil {
				log.Printf("Error handling reservation.confirmed: %s", err)
			}
		case entities.MessageTypeReservationCancelled:
			var reservation entities.CancelledNotiReservation
			if err := json.Unmarshal(eventData.Data, &reservation); err != nil {
				log.Printf("Error unmarshaling reservation.cancelled payload: %s", err)
				continue
			}

			if err := h.domain.ReservationCancelledEvent(ctx, reservation); err != nil {
				log.Printf("Error handling reservation.cancelled: %s", err)
			}
		default:
			log.Printf("Unknown message type: %s", eventData.Type)
			continue
		}
	}

	return nil
}
