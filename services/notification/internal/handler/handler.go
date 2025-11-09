package handler

import (
	"context"

	"github.com/cp-rektmart/aconcert-microservice/notification/internal/domain"
)

type Handler struct {
	domain *domain.Domain
}

func New(domain *domain.Domain) *Handler {
	return &Handler{
		domain: domain,
	}
}

func (h *Handler) HandleEventCreated(ctx context.Context) error {
	panic("unimplemented")
}

func (h *Handler) HandleEventUpdated(ctx context.Context) error {
	panic("unimplemented")
}

func (h *Handler) HandleReservationConfirmed(ctx context.Context) error {
	panic("unimplemented")
}

func (h *Handler) HandleReservationCancelled(ctx context.Context) error {
	panic("unimplemented")
}
