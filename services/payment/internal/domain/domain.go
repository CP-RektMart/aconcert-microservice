package domain

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/payment/internal/repository"
	"github.com/stripe/stripe-go/v83"
)

type Domain struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Domain {
	return &Domain{
		repo: repo,
	}
}

func (d *Domain) ProcessPayment(ctx context.Context, event stripe.Event) error {
	if event.Type == stripe.EventTypeCheckoutSessionCompleted {
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return errors.Wrap(err, "failed to parse stripe webhook")
		}

		reservation, err := d.repo.GetReservationByStripeSessionID(ctx, session.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get reservation by stripe session id")
		}

		if err := d.repo.ConfirmPayment(ctx, reservation.ID); err != nil {
			return errors.Wrap(err, "failed to confirm payment")
		}
	}

	return nil
}
