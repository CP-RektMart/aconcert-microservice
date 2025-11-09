package domains

import (
	"context"

	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/repositories"
)

type ReserveDomain interface {
	CreateReservation(ctx context.Context, req *reservationpb.CreateReservationRequest) (*reservationpb.CreateReservationResponse, error)
	DeleteReservation(ctx context.Context, req *reservationpb.DeleteReservationRequest) (*reservationpb.DeleteReservationResponse, error)
	GetReservation(ctx context.Context, req *reservationpb.GetReservationRequest) (*reservationpb.GetReservationResponse, error)
	ListReservation(ctx context.Context, req *reservationpb.ListReservationRequest) (*reservationpb.ListReservationResponse, error)
	ConfirmReservation(ctx context.Context, req *reservationpb.ConfirmReservationRequest) (*reservationpb.ConfirmReservationResponse, error)
}

type ReserveDomainImpl struct {
	reservationpb.UnimplementedReservationServiceServer
	repo repositories.ReservationRepository
}

func New(repo repositories.ReservationRepository) *ReserveDomainImpl {
	return &ReserveDomainImpl{
		repo: repo,
	}
}
