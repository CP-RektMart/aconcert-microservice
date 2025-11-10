package repositories

import (
	"context"
	"time"

	db "github.com/cp-rektmart/aconcert-microservice/reservation/db/codegen"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/entities"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type SeatInfo struct {
	ZoneNumber int32
	RowNumber  int32
	ColNumber  int32
}

type SeatStatusInfo struct {
	ZoneNumber int32
	RowNumber  int32
	ColNumber  int32
	Status     string // "PENDING" or "RESERVED"
}

type ReservationRepository interface {
	// redis
	CreateReservationTemp(ctx context.Context, userID, reservationID string, ttl time.Duration) error
	GetReservationTimeLeft(ctx context.Context, userID, reservationID string) (time.Duration, error)
	DeleteReservationTemp(ctx context.Context, userID, reservationID string) error
	CheckSeatAvailable(ctx context.Context, eventID string, seat SeatInfo) (bool, error)
	SetSeatReserved(ctx context.Context, eventID string, seat SeatInfo, reservationID string, ttl time.Duration) error
	SetSeatTempReserved(ctx context.Context, eventID string, seat SeatInfo, reservationID string, ttl time.Duration) error
	DeleteSeatReservation(ctx context.Context, eventID string, seat SeatInfo) error
	CacheReservationSeats(ctx context.Context, reservationID string, seats []SeatInfo, ttl time.Duration) error
	GetReservationSeats(ctx context.Context, reservationID string) ([]SeatInfo, error)
	DeleteReservationSeats(ctx context.Context, reservationID string) error
	GetAllEventSeats(ctx context.Context, eventID string) ([]SeatStatusInfo, error)

	// pub/sub - redis
	publishSeatUpdate(ctx context.Context, eventID string, seat SeatInfo, status entities.SeatStatus)

	// redis event
	StartExpirationListener(ctx context.Context)

	// db
	GetReservation(ctx context.Context, id string) (*db.Reservation, error)
	ListReservationsByUserID(ctx context.Context, userID string) ([]db.Reservation, error)
	CreateReservation(ctx context.Context, reservationID, userID, eventID, status, stripeSessionID string, totalPrice float64) (*db.Reservation, error)
	UpdateReservationStatus(ctx context.Context, id, status string) (*db.Reservation, error)
	DeleteReservation(ctx context.Context, id string) error
	CreateTicket(ctx context.Context, eventID, reservationID string, seat SeatInfo) (*db.Ticket, error)
	CreateTickets(ctx context.Context, eventID, reservationID string, seats []SeatInfo) ([]db.Ticket, error)
	CreateTicketsWithTransaction(ctx context.Context, eventID, reservationID string, seats []SeatInfo) ([]db.Ticket, error)
	GetTicketsByReservation(ctx context.Context, reservationID string) ([]db.Ticket, error)
	GetReservationBySessionId(ctx context.Context, sessionID string) (*db.Reservation, error)
}

type ReservationImpl struct {
	db          *db.Queries
	pool        *pgxpool.Pool
	redisClient *redis.Client
}

func NewReservationRepository(db *db.Queries, pool *pgxpool.Pool, redisClient *redis.Client) *ReservationImpl {
	return &ReservationImpl{
		db:          db,
		pool:        pool,
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
