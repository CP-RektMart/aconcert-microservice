package repositories

import (
	"context"
	"fmt"

	db "github.com/cp-rektmart/aconcert-microservice/reservation/db/codegen"
)

func (r *ReservationImpl) GetReservation(ctx context.Context, id string) (*db.Reservation, error) {
	uuid := stringToUUID(id)
	reservation, err := r.db.GetReservation(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationImpl) CreateReservation(ctx context.Context, userID, eventID, status string) (*db.Reservation, error) {
	params := db.CreateReservationParams{
		UserID:  stringToUUID(userID),
		EventID: stringToUUID(eventID),
		Status:  status,
	}

	reservation, err := r.db.CreateReservation(ctx, params)
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationImpl) UpdateReservationStatus(ctx context.Context, id, status string) (*db.Reservation, error) {
	params := db.UpdateReservationStatusParams{
		ID:     stringToUUID(id),
		Status: status,
	}

	reservation, err := r.db.UpdateReservationStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationImpl) DeleteReservation(ctx context.Context, id string) error {
	uuid := stringToUUID(id)
	return r.db.DeleteReservation(ctx, uuid)
}

func (r *ReservationImpl) CreateTicket(ctx context.Context, reservationID string, seat SeatInfo) (*db.Ticket, error) {
	params := db.CreateTicketParams{
		ReservationID: stringToUUID(reservationID),
		ZoneNumber:    seat.ZoneNumber,
		RowNumber:     seat.RowNumber,
		ColNumber:     seat.ColNumber,
	}

	ticket, err := r.db.CreateTicket(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ReservationImpl) CreateTickets(ctx context.Context, reservationID string, seats []SeatInfo) ([]db.Ticket, error) {
	tickets := make([]db.Ticket, 0, len(seats))

	for _, seat := range seats {
		ticket, err := r.CreateTicket(ctx, reservationID, seat)
		if err != nil {
			return nil, fmt.Errorf("failed to create ticket for seat %+v: %w", seat, err)
		}
		tickets = append(tickets, *ticket)
	}

	return tickets, nil
}

func (r *ReservationImpl) GetTicketsByReservation(ctx context.Context, reservationID string) ([]db.Ticket, error) {
	uuid := stringToUUID(reservationID)
	return r.db.ListTicketsByReservationID(ctx, uuid)
}
