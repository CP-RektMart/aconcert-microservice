package repositories

import (
	"context"
	"time"

	db "github.com/cp-rektmart/aconcert-microservice/reservation/db/codegen"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type SeatInfo struct {
	ZoneNumber int32
	RowNumber  int32
	ColNumber  int32
}

type ReservationRepository interface {
	CreateReservationTemp(ctx context.Context, userID, reservationID string, ttl time.Duration) error
	GetReservationTimeLeft(ctx context.Context, userID, reservationID string) (time.Duration, error)
	DeleteReservationTemp(ctx context.Context, userID, reservationID string) error
	CheckSeatAvailable(ctx context.Context, eventID string, seat SeatInfo) (bool, error)
	SetSeatReserved(ctx context.Context, eventID string, seat SeatInfo, reservationID string, ttl time.Duration) error
	DeleteSeatReservation(ctx context.Context, eventID string, seat SeatInfo) error
	CacheReservationSeats(ctx context.Context, reservationID string, seats []SeatInfo, ttl time.Duration) error
	GetReservationSeats(ctx context.Context, reservationID string) ([]SeatInfo, error)
	DeleteReservationSeats(ctx context.Context, reservationID string) error

	GetReservation(ctx context.Context, id string) (*db.Reservation, error)
	CreateReservation(ctx context.Context, userID, eventID, status string) (*db.Reservation, error)
	UpdateReservationStatus(ctx context.Context, id, status string) (*db.Reservation, error)
	DeleteReservation(ctx context.Context, id string) error
	CreateTicket(ctx context.Context, reservationID string, seat SeatInfo) (*db.Ticket, error)
	CreateTickets(ctx context.Context, reservationID string, seats []SeatInfo) ([]db.Ticket, error)
	GetTicketsByReservation(ctx context.Context, reservationID string) ([]db.Ticket, error)
}

type ReservationImpl struct {
	db          *db.Queries
	redisClient *redis.Client
}

func NewReservationRepository(db *db.Queries, redisClient *redis.Client) *ReservationImpl {
	return &ReservationImpl{
		db:          db,
		redisClient: redisClient,
	}
}

func stringToUUID(s string) pgtype.UUID {
	var uuid pgtype.UUID
	if err := uuid.Scan(s); err != nil {
		return pgtype.UUID{Valid: false}
	}
	return uuid
}

func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return uuid.String()
}
