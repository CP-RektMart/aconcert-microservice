package repositories

import (
	"context"
	"fmt"

	db "github.com/cp-rektmart/aconcert-microservice/reservation/db/codegen"
	"github.com/jackc/pgx/v5"
)

func (r *ReservationImpl) GetReservation(ctx context.Context, id string) (*db.Reservation, error) {
	uuid := stringToUUID(id)
	reservation, err := r.db.GetReservation(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationImpl) CreateReservation(ctx context.Context, reservationID string, userID, eventID, status, stripeSessionID string) (*db.Reservation, error) {
	params := db.CreateReservationParams{
		ID:              stringToUUID(reservationID),
		UserID:          stringToUUID(userID),
		EventID:         stringToUUID(eventID),
		Status:          status,
		StripeSessionID: stripeSessionID,
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

func (r *ReservationImpl) CreateTicket(ctx context.Context, eventID, reservationID string, seat SeatInfo) (*db.Ticket, error) {
	params := db.CreateTicketParams{
		ReservationID: stringToUUID(reservationID),
		ZoneNumber:    seat.ZoneNumber,
		RowNumber:     seat.RowNumber,
		ColNumber:     seat.ColNumber,
		EventID:       stringToUUID(eventID),
	}

	ticket, err := r.db.CreateTicket(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ReservationImpl) CreateTickets(ctx context.Context, eventID, reservationID string, seats []SeatInfo) ([]db.Ticket, error) {
	tickets := make([]db.Ticket, 0, len(seats))

	for _, seat := range seats {
		ticket, err := r.CreateTicket(ctx, eventID, reservationID, seat)
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

func (r *ReservationImpl) ListReservationsByUserID(ctx context.Context, userID string) ([]db.Reservation, error) {
	uuid := stringToUUID(userID)
	return r.db.ListReservationsByUserID(ctx, uuid)
}

// CreateTicketsWithTransaction creates tickets within a database transaction
// It checks seat availability for the event and creates tickets atomically
func (r *ReservationImpl) CreateTicketsWithTransaction(ctx context.Context, eventID, reservationID string, seats []SeatInfo) ([]db.Ticket, error) {
	if r.pool == nil {
		// Fallback to non-transactional if pool is not available
		return r.CreateTickets(ctx, eventID, reservationID, seats)
	}

	var tickets []db.Ticket

	// Start a database transaction
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.db.WithTx(tx)
		eventUUID := stringToUUID(eventID)
		reservationUUID := stringToUUID(reservationID)

		// Check availability for all seats first
		for _, seat := range seats {
			params := db.CheckSeatAvailabilityForEventParams{
				EventID:    eventUUID,
				ZoneNumber: seat.ZoneNumber,
				RowNumber:  seat.RowNumber,
				ColNumber:  seat.ColNumber,
			}

			isTaken, err := queries.CheckSeatAvailabilityForEvent(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to check seat availability for zone=%d row=%d col=%d: %w",
					seat.ZoneNumber, seat.RowNumber, seat.ColNumber, err)
			}

			if isTaken {
				return fmt.Errorf("seat already taken: zone=%d row=%d col=%d",
					seat.ZoneNumber, seat.RowNumber, seat.ColNumber)
			}
		}

		// If all seats are available, create tickets
		for _, seat := range seats {
			params := db.CreateTicketParams{
				ReservationID: reservationUUID,
				ZoneNumber:    seat.ZoneNumber,
				RowNumber:     seat.RowNumber,
				ColNumber:     seat.ColNumber,
				EventID:       eventUUID,
			}

			ticket, err := queries.CreateTicket(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to create ticket for seat zone=%d row=%d col=%d: %w",
					seat.ZoneNumber, seat.RowNumber, seat.ColNumber, err)
			}

			tickets = append(tickets, ticket)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tickets, nil
}
