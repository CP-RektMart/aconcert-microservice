package reservation

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/google/uuid"
)

type ReservationService struct {
	client reservationpb.ReservationServiceClient
}

func NewService(client reservationpb.ReservationServiceClient) *ReservationService {
	return &ReservationService{
		client: client,
	}
}

// TransformSeatToDTO transforms a protobuf Seat to DTO
func (s *ReservationService) TransformSeatToDTO(seat *reservationpb.Seat) dto.SeatDTO {
	return dto.SeatDTO{
		ZoneNumber: seat.ZoneNumber,
		Row:        seat.Row,
		Column:     seat.Column,
		Price:      seat.Price,
	}
}

// TransformSeatToProto transforms a DTO Seat to protobuf
func (s *ReservationService) TransformSeatToProto(seat dto.SeatDTO) *reservationpb.Seat {
	return &reservationpb.Seat{
		ZoneNumber: seat.ZoneNumber,
		Row:        seat.Row,
		Column:     seat.Column,
		Price:      seat.Price,
	}
}

// CreateReservation creates a new reservation
func (s *ReservationService) CreateReservation(ctx context.Context, req *dto.CreateReservationRequest, userID uuid.UUID) (string, error) {
	seats := make([]*reservationpb.Seat, 0, len(req.Seats))
	for _, seat := range req.Seats {
		seats = append(seats, s.TransformSeatToProto(seat))
	}

	transUserID := userID.String()

	response, err := s.client.CreateReservation(ctx, &reservationpb.CreateReservationRequest{
		UserId:     transUserID,
		EventId:    req.EventID,
		TotalPrice: req.TotalPrice,
		Seats:      seats,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create reservation")
	}

	return response.Id, nil
}

// DeleteReservation deletes a reservation
func (s *ReservationService) DeleteReservation(ctx context.Context, req *dto.DeleteReservationRequest) (string, error) {
	response, err := s.client.DeleteReservation(ctx, &reservationpb.DeleteReservationRequest{
		Id: req.ID,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to delete reservation")
	}

	return response.Id, nil
}

// GetReservation gets a reservation by ID
func (s *ReservationService) GetReservation(ctx context.Context, req *dto.GetReservationRequest) (dto.GetReservationResponse, error) {
	response, err := s.client.GetReservation(ctx, &reservationpb.GetReservationRequest{
		Id: req.ID,
	})
	if err != nil {
		return dto.GetReservationResponse{}, errors.Wrap(err, "failed to get reservation")
	}

	seats := make([]dto.SeatDTO, 0, len(response.Seats))
	for _, seat := range response.Seats {
		seats = append(seats, s.TransformSeatToDTO(seat))
	}

	return dto.GetReservationResponse{
		ID:                 response.Id,
		UserID:             response.UserId,
		EventID:            response.EventId,
		TotalPrice:         response.TotalPrice,
		Seats:              seats,
		StripeClientSecret: response.StripeClientSecret,
		TimeLeft:           *response.TimeLeft,
	}, nil
}

// ListReservation lists all reservations for a user
func (s *ReservationService) ListReservation(ctx context.Context, req *dto.ListReservationRequest) ([]dto.ReservationDTO, error) {
	response, err := s.client.ListReservation(ctx, &reservationpb.ListReservationRequest{
		UserId: req.UserID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list reservations")
	}

	reservations := make([]dto.ReservationDTO, 0, len(response.Reservation))
	for _, reservation := range response.Reservation {
		seats := make([]dto.SeatDTO, 0, len(reservation.Seats))
		for _, seat := range reservation.Seats {
			seats = append(seats, s.TransformSeatToDTO(seat))
		}

		reservations = append(reservations, dto.ReservationDTO{
			ID:         "", // ID not available in protobuf Reservation message
			UserID:     reservation.UserId,
			EventID:    reservation.EventId,
			TotalPrice: reservation.TotalPrice,
			Seats:      seats,
		})
	}

	return reservations, nil
}

// ConfirmReservation confirms a reservation
func (s *ReservationService) ConfirmReservation(ctx context.Context, req *dto.ConfirmReservationRequest) (dto.ConfirmReservationResponse, error) {
	response, err := s.client.ConfirmReservation(ctx, &reservationpb.ConfirmReservationRequest{
		Id: req.ID,
	})
	if err != nil {
		return dto.ConfirmReservationResponse{}, errors.Wrap(err, "failed to confirm reservation")
	}

	return dto.ConfirmReservationResponse{
		ID:      response.Id,
		Success: response.Success,
		Message: response.Message,
	}, nil
}
