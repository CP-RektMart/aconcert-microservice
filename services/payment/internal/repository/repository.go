package repository

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/payment/internal/entities"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
)

type Repository struct {
	reservationClient reservationpb.ReservationServiceClient
}

func NewRepository(reservationClient reservationpb.ReservationServiceClient) *Repository {
	return &Repository{
		reservationClient: reservationClient,
	}
}

func (r *Repository) GetReservationByStripeSessionID(ctx context.Context, sessionID string) (entities.Reservation, error) {
	response, err := r.reservationClient.GetReservationByStripeSessionID(ctx, &reservationpb.GetReservationByStripeSessionIDRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return entities.Reservation{}, errors.Wrap(err, "failed to get reservation by stripe session ID")
	}

	seats := make([]entities.Seat, len(response.Seats))
	for i, seat := range response.Seats {
		seats[i] = entities.Seat{
			ZoneNumber: int(seat.ZoneNumber),
			Price:      seat.Price,
			Row:        int(seat.Row),
			Column:     int(seat.Column),
		}
	}

	return entities.Reservation{
		ID:         response.Id,
		UserID:     response.UserId,
		EventID:    response.EventId,
		TotalPrice: response.TotalPrice,
		Seats:      seats,
	}, nil
}

func (r *Repository) ConfirmPayment(ctx context.Context, reservationID string) error {
	response, err := r.reservationClient.ConfirmReservation(ctx, &reservationpb.ConfirmReservationRequest{
		Id: reservationID,
	})

	if err != nil {
		return errors.Wrap(err, "failed to confirm reservation")
	}

	if !response.Success {
		return errors.New("NOT_SUCCESSFUL")
	}

	return nil
}
