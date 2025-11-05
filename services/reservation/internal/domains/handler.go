package domains

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cp-rektmart/aconcert-microservice/pkg/apperror"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/repositories"
)

const (
	SafetyBuffer    = 30 * time.Second
	ReservationTTL  = 5*time.Minute + SafetyBuffer
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
)

func (r *ReserveDomainImpl) CreateReservation(ctx context.Context, req *reservationpb.Reservation) (*reservationpb.CreateReservationResponse, error) {
	if err := validateReservationRequest(req); err != nil {
		return nil, err
	}

	seats := convertSeatsToSeatInfo(req.GetSeats())

	for _, seat := range seats {
		available, err := r.repo.CheckSeatAvailable(ctx, req.GetEventId(), seat)
		if err != nil {
			logger.ErrorContext(ctx, "check seat availability failed", slog.Any("error", err))
			return nil, apperror.Internal("failed to check seat availability", err)
		}
		if !available {
			logger.WarnContext(ctx, "seat already reserved", slog.Int("zone", int(seat.ZoneNumber)), slog.Int("row", int(seat.RowNumber)), slog.Int("col", int(seat.ColNumber)))
			return nil, apperror.BadRequest("seat already reserved", nil)
		}
	}

	reservationID := uuid.New().String()

	_, err := r.repo.CreateReservation(ctx, req.GetUserId(), req.GetEventId(), StatusPending)
	if err != nil {
		logger.ErrorContext(ctx, "create reservation failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to create reservation", err)
	}

	if err := r.repo.CreateReservationTemp(ctx, req.GetUserId(), reservationID, ReservationTTL); err != nil {
		logger.ErrorContext(ctx, "cache reservation failed", slog.Any("error", err))
		r.repo.DeleteReservation(ctx, reservationID)
		return nil, apperror.Internal("failed to cache reservation", err)
	}

	if err := r.repo.CacheReservationSeats(ctx, reservationID, seats, ReservationTTL); err != nil {
		logger.ErrorContext(ctx, "cache seats failed", slog.Any("error", err))
		r.repo.DeleteReservationTemp(ctx, req.GetUserId(), reservationID)
		r.repo.DeleteReservation(ctx, reservationID)
		return nil, apperror.Internal("failed to cache seats", err)
	}

	for _, seat := range seats {
		if err := r.repo.SetSeatReserved(ctx, req.GetEventId(), seat, reservationID, ReservationTTL); err != nil {
			logger.ErrorContext(ctx, "reserve seat failed", slog.Any("error", err))
			rollbackReservation(ctx, r.repo, req.GetUserId(), req.GetEventId(), reservationID, seats)
			return nil, apperror.Internal("failed to reserve seat", err)
		}
	}

	logger.InfoContext(ctx, "reservation created", slog.String("reservationID", reservationID), slog.String("userID", req.GetUserId()))
	return &reservationpb.CreateReservationResponse{
		Id: reservationID,
	}, nil
}

func (r *ReserveDomainImpl) DeleteReservation(ctx context.Context, req *reservationpb.DeleteReservationRequest) (*reservationpb.DeleteReservationResponse, error) {
	reservationID := req.GetId()
	userID := ""
	eventID := ""

	reservation, err := r.repo.GetReservation(ctx, reservationID)
	if err != nil {
		logger.ErrorContext(ctx, "get reservation failed", slog.Any("error", err))
		return nil, apperror.NotFound("reservation not found", err)
	}
	userID = pgUUIDToString(reservation.UserID)
	eventID = pgUUIDToString(reservation.EventID)

	timeLeft, err := r.repo.GetReservationTimeLeft(ctx, userID, reservationID)
	if err != nil || timeLeft <= 0 {
		logger.WarnContext(ctx, "reservation not found or expired", slog.String("reservationID", reservationID))
		return nil, apperror.NotFound("reservation not found or expired", err)
	}

	if timeLeft < SafetyBuffer {
		logger.WarnContext(ctx, "reservation expiring soon", slog.String("reservationID", reservationID), slog.Duration("timeLeft", timeLeft))
		return nil, apperror.BadRequest("reservation expiring soon, cannot cancel", nil)
	}

	if err := r.repo.DeleteReservationTemp(ctx, userID, reservationID); err != nil {
		logger.ErrorContext(ctx, "delete reservation cache failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to delete reservation cache", err)
	}

	if err := r.repo.DeleteReservation(ctx, reservationID); err != nil {
		logger.ErrorContext(ctx, "delete reservation failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to delete reservation", err)
	}

	seats, _ := r.repo.GetReservationSeats(ctx, reservationID)
	for _, seat := range seats {
		r.repo.DeleteSeatReservation(ctx, eventID, seat)
	}
	r.repo.DeleteReservationSeats(ctx, reservationID)

	logger.InfoContext(ctx, "reservation cancelled", slog.String("reservationID", reservationID))
	return &reservationpb.DeleteReservationResponse{
		Id: reservationID,
	}, nil
}

func (r *ReserveDomainImpl) GetReservation(ctx context.Context, req *reservationpb.GetReservationRequest) (*reservationpb.GetReservationResponse, error) {
	reservation, err := r.repo.GetReservation(ctx, req.GetId())
	if err != nil {
		logger.ErrorContext(ctx, "get reservation failed", slog.Any("error", err))
		return nil, apperror.NotFound("reservation not found", err)
	}

	tickets, err := r.repo.GetTicketsByReservation(ctx, req.GetId())
	if err != nil {
		logger.ErrorContext(ctx, "get tickets failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to get tickets", err)
	}

	seats := make([]*reservationpb.Seat, len(tickets))
	for i, ticket := range tickets {
		seats[i] = &reservationpb.Seat{
			ZoneNumber: ticket.ZoneNumber,
			Row:        ticket.RowNumber,
			Column:     ticket.ColNumber,
		}
	}

	return &reservationpb.GetReservationResponse{
		Id:      req.GetId(),
		UserId:  pgUUIDToString(reservation.UserID),
		EventId: pgUUIDToString(reservation.EventID),
		Seats:   seats,
	}, nil
}

func (r *ReserveDomainImpl) ListReservation(ctx context.Context, req *reservationpb.ListReservationRequest) (*reservationpb.ListReservationResponse, error) {
	return &reservationpb.ListReservationResponse{}, nil
}

func (r *ReserveDomainImpl) ConfirmReservation(ctx context.Context, userID, reservationID string) error {
	timeLeft, err := r.repo.GetReservationTimeLeft(ctx, userID, reservationID)
	if err != nil || timeLeft <= 0 {
		logger.WarnContext(ctx, "reservation not found or expired", slog.String("reservationID", reservationID))
		return apperror.NotFound("reservation not found or expired", err)
	}

	if timeLeft < SafetyBuffer {
		logger.WarnContext(ctx, "reservation expiring soon", slog.String("reservationID", reservationID), slog.Duration("timeLeft", timeLeft))
		return apperror.BadRequest("reservation expiring soon, please create new", nil)
	}

	reservation, err := r.repo.GetReservation(ctx, reservationID)
	if err != nil {
		logger.ErrorContext(ctx, "get reservation failed", slog.Any("error", err))
		return apperror.NotFound("reservation not found", err)
	}
	eventID := pgUUIDToString(reservation.EventID)

	seats, err := r.repo.GetReservationSeats(ctx, reservationID)
	if err != nil {
		logger.ErrorContext(ctx, "get cached seats failed", slog.Any("error", err))
		return apperror.Internal("failed to get reservation seats", err)
	}

	if _, err := r.repo.CreateTickets(ctx, reservationID, seats); err != nil {
		logger.ErrorContext(ctx, "create tickets failed", slog.Any("error", err))
		return apperror.Internal("failed to create tickets", err)
	}

	if _, err := r.repo.UpdateReservationStatus(ctx, reservationID, StatusConfirmed); err != nil {
		logger.ErrorContext(ctx, "confirm reservation failed", slog.Any("error", err))
		return apperror.Internal("failed to confirm reservation", err)
	}

	if err := r.repo.DeleteReservationTemp(ctx, userID, reservationID); err != nil {
		logger.ErrorContext(ctx, "cleanup temp reservation failed", slog.Any("error", err))
		return apperror.Internal("failed to cleanup temp reservation", err)
	}

	for _, seat := range seats {
		r.repo.DeleteSeatReservation(ctx, eventID, seat)
	}
	r.repo.DeleteReservationSeats(ctx, reservationID)

	logger.InfoContext(ctx, "reservation confirmed", slog.String("reservationID", reservationID))
	return nil
}

func validateReservationRequest(req *reservationpb.Reservation) error {
	if req.GetUserId() == "" {
		return apperror.BadRequest("user ID required", nil)
	}
	if req.GetEventId() == "" {
		return apperror.BadRequest("event ID required", nil)
	}
	if len(req.GetSeats()) == 0 {
		return apperror.BadRequest("at least one seat required", nil)
	}
	return nil
}

func convertSeatsToSeatInfo(seats []*reservationpb.Seat) []repositories.SeatInfo {
	result := make([]repositories.SeatInfo, len(seats))
	for i, seat := range seats {
		result[i] = repositories.SeatInfo{
			ZoneNumber: seat.GetZoneNumber(),
			RowNumber:  seat.GetRow(),
			ColNumber:  seat.GetColumn(),
		}
	}
	return result
}

func rollbackReservation(ctx context.Context, repo repositories.ReservationRepository, userID, eventID, reservationID string, seats []repositories.SeatInfo) {
	repo.DeleteReservationTemp(ctx, userID, reservationID)
	repo.DeleteReservationSeats(ctx, reservationID)
	repo.DeleteReservation(ctx, reservationID)
	for _, seat := range seats {
		repo.DeleteSeatReservation(ctx, eventID, seat)
	}
}

func pgUUIDToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return uuid.String()
}
