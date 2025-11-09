package domains

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cp-rektmart/aconcert-microservice/pkg/apperror"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/repositories"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

const (
	SafetyBuffer   = 30 * time.Second
	ResevationMax  = 5 * time.Minute
	ReservationTTL = ResevationMax + SafetyBuffer
)

func (r *ReserveDomainImpl) CreateReservation(ctx context.Context, req *reservationpb.CreateReservationRequest) (*reservationpb.CreateReservationResponse, error) {
	var err error
	var reservationID string
	var seats []repositories.SeatInfo

	// rollback purpose
	defer func() {
		if err != nil {
			rollbackReservation(ctx, r.repo, req.GetUserId(), req.GetEventId(), reservationID, seats)
		}
	}()

	if err := validateReservationRequest(req); err != nil {
		logger.ErrorContext(ctx, "check seat availability failed", slog.Any("error", err))
		return nil, err
	}

	seats = convertSeatsToSeatInfo(req.GetSeats())

	// quickly check by the cache
	for _, seat := range seats {
		available, err := r.repo.CheckSeatAvailable(ctx, req.GetEventId(), seat)
		if err != nil {
			return nil, apperror.Internal("failed to check seat availability", err)
		}
		if !available {
			return nil, apperror.BadRequest("seat already reserved", nil)
		}
	}

	reservationID = uuid.New().String()
	// cache the reservation
	if err := r.repo.CreateReservationTemp(ctx, req.GetUserId(), reservationID, ReservationTTL); err != nil {
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
		if err := r.repo.SetSeatTempReserved(ctx, req.GetEventId(), seat, reservationID, ReservationTTL); err != nil {
			logger.ErrorContext(ctx, "reserve seat failed", slog.Any("error", err))
			rollbackReservation(ctx, r.repo, req.GetUserId(), req.GetEventId(), reservationID, seats)
			return nil, apperror.Internal("failed to reserve seat", err)
		}
	}

	stripe.Key = r.stripe.SecretKey
	params := &stripe.CheckoutSessionParams{
		UIMode: stripe.String("embedded"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("thb"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("your reservation checkout"),
					},
					// handle in satang unit
					UnitAmount: stripe.Int64(int64(req.GetTotalPrice() * 100)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		ReturnURL:         stripe.String(r.stripe.ReturnURL + "/?session_id={CHECKOUT_SESSION_ID}"),
		ClientReferenceID: stripe.String(fmt.Sprintf("%s", reservationID)),
	}

	session, err := session.New(params)
	if err != nil {
		return nil, apperror.Internal("Failed to create Stripe session", err)
	}

	_, err = r.repo.CreateReservation(ctx, reservationID, req.GetUserId(), req.GetEventId(), string(entities.Pending), session.ID, req.GetTotalPrice())
	if err != nil {
		return nil, apperror.Internal("Failed to create reservation", err)
	}

	logger.InfoContext(ctx, "reservation created", slog.String("reservationID", reservationID), slog.String("userID", req.GetUserId()))
	return &reservationpb.CreateReservationResponse{
		Id: reservationID,
	}, nil
}

func (r *ReserveDomainImpl) DeleteReservation(ctx context.Context, req *reservationpb.DeleteReservationRequest) (*reservationpb.DeleteReservationResponse, error) {
	reservationID := req.GetId()

	reservation, err := r.repo.GetReservation(ctx, reservationID)
	if err != nil {
		return nil, apperror.NotFound("reservation not found", err)
	}

	reservationID = pgUUIDToString(reservation.ID)
	userID := pgUUIDToString(reservation.UserID)
	eventID := pgUUIDToString(reservation.EventID)

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

	var seats []*reservationpb.Seat
	var timeLeft float64

	switch reservation.Status {
	case string(entities.Pending):
		// Handle pending status
		tmpSeats, err := r.repo.GetReservationSeats(ctx, req.GetId())
		if err != nil {
			logger.ErrorContext(ctx, "get reservation seats failed", slog.Any("error", err))
			return nil, apperror.Internal("failed to get reservation seats", err)
		}
		for _, seat := range tmpSeats {
			seats = append(seats, &reservationpb.Seat{
				ZoneNumber: seat.ZoneNumber,
				Row:        seat.RowNumber,
				Column:     seat.ColNumber,
			})
		}
		rtTime, err := r.repo.GetReservationTimeLeft(ctx, reservation.UserID.String(), reservation.ID.String())
		if err != nil {
			return nil, apperror.Internal("failed to get reservation time left", err)
		}

		// Convert time.Duration to float64 (seconds)
		timeLeftSeconds := rtTime.Seconds()
		timeLeft = min(timeLeftSeconds-SafetyBuffer.Seconds(), ResevationMax.Seconds())

	case string(entities.Confirmed):
		// Handle confirmed status
		tickets, err := r.repo.GetTicketsByReservation(ctx, req.GetId())
		if err != nil {
			logger.ErrorContext(ctx, "get tickets failed", slog.Any("error", err))
			return nil, apperror.Internal("failed to get tickets", err)
		}
		for _, ticket := range tickets {
			seats = append(seats, &reservationpb.Seat{
				ZoneNumber: ticket.ZoneNumber,
				Row:        ticket.RowNumber,
				Column:     ticket.ColNumber,
			})
		}
	case string(entities.Cancelled):
		break
	default:
		break
	}

	session, err := session.Get(reservation.StripeSessionID, &stripe.CheckoutSessionParams{})
	if err != nil {
		return nil, apperror.Internal("Failed to get Stripe Client Secret", err)
	}

	return &reservationpb.GetReservationResponse{
		Id:                 req.GetId(),
		UserId:             pgUUIDToString(reservation.UserID),
		EventId:            pgUUIDToString(reservation.EventID),
		Seats:              seats,
		TimeLeft:           &timeLeft,
		StripeClientSecret: session.ClientSecret,
		Status:             reservation.Status,
		TotalPrice:         reservation.TotalPrice,
	}, nil
}

func (r *ReserveDomainImpl) ListReservation(ctx context.Context, req *reservationpb.ListReservationRequest) (*reservationpb.ListReservationResponse, error) {
	userID := req.GetUserId()

	if userID == "" {
		logger.WarnContext(ctx, "list reservation missing userID")
		return nil, apperror.BadRequest("user ID required", nil)
	}

	reservations, err := r.repo.ListReservationsByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "list reservations failed", slog.Any("error", err), slog.String("userID", userID))
		return nil, apperror.Internal("failed to list reservations", err)
	}

	logger.InfoContext(ctx, "reservations listed", slog.String("userID", userID), slog.Int("count", len(reservations)))
	return &reservationpb.ListReservationResponse{}, nil
}
func (r *ReserveDomainImpl) ConfirmReservation(ctx context.Context, req *reservationpb.ConfirmReservationRequest) (*reservationpb.ConfirmReservationResponse, error) {
	reservationID := req.GetId()

	reservation, err := r.repo.GetReservation(ctx, reservationID)
	if err != nil {
		logger.ErrorContext(ctx, "get reservation failed", slog.Any("error", err))
		return nil, apperror.NotFound("reservation not found", err)
	}

	timeLeft, err := r.repo.GetReservationTimeLeft(ctx, reservation.UserID.String(), reservationID)
	if err != nil || timeLeft <= 0 {
		logger.WarnContext(ctx, "reservation not found or expired", slog.String("reservationID", reservationID))
		return nil, apperror.NotFound("reservation not found or expired", err)
	}

	eventID := pgUUIDToString(reservation.EventID)

	seats, err := r.repo.GetReservationSeats(ctx, reservationID)
	if err != nil {
		logger.ErrorContext(ctx, "get cached seats failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to get reservation seats", err)
	}

	if _, err := r.repo.CreateTicketsWithTransaction(ctx, eventID, reservationID, seats); err != nil {
		logger.ErrorContext(ctx, "create tickets failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to create tickets", err)
	}

	for _, seat := range seats {
		if err := r.repo.SetSeatReserved(ctx, reservation.EventID.String(), seat, reservationID, ReservationTTL); err != nil {
			logger.ErrorContext(ctx, "reserve seat failed", slog.Any("error", err))
			rollbackReservation(ctx, r.repo, reservation.UserID.String(), reservation.EventID.String(), reservationID, seats)
			return nil, apperror.Internal("failed to reserve seat", err)
		}
	}

	if _, err := r.repo.UpdateReservationStatus(ctx, reservationID, string(entities.Confirmed)); err != nil {
		logger.ErrorContext(ctx, "confirm reservation failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to confirm reservation", err)
	}

	if err := r.repo.DeleteReservationTemp(ctx, reservation.UserID.String(), reservationID); err != nil {
		logger.ErrorContext(ctx, "cleanup temp reservation failed", slog.Any("error", err))
		return nil, apperror.Internal("failed to cleanup temp reservation", err)
	}

	logger.InfoContext(ctx, "reservation confirmed", slog.String("reservationID", reservationID))
	return &reservationpb.ConfirmReservationResponse{
		Id:      reservationID,
		Success: true,
		Message: "Reservation confirmed",
	}, nil
}

func validateReservationRequest(req *reservationpb.CreateReservationRequest) error {
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
