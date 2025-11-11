package domains

import (
	"context"

	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/reservation/config"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/repositories"
)

type ReserveDomain interface {
	CreateReservation(ctx context.Context, req *reservationpb.CreateReservationRequest) (*reservationpb.CreateReservationResponse, error)
	DeleteReservation(ctx context.Context, req *reservationpb.DeleteReservationRequest) (*reservationpb.DeleteReservationResponse, error)
	GetReservation(ctx context.Context, req *reservationpb.GetReservationRequest) (*reservationpb.GetReservationResponse, error)
	ListReservation(ctx context.Context, req *reservationpb.ListReservationRequest) (*reservationpb.ListReservationResponse, error)
	ConfirmReservation(ctx context.Context, req *reservationpb.ConfirmReservationRequest) (*reservationpb.ConfirmReservationResponse, error)
	GetReservationByStripeSessionID(ctx context.Context, req *reservationpb.GetReservationByStripeSessionIDRequest) (*reservationpb.GetReservationResponse, error)
}

type ReserveDomainImpl struct {
	reservationpb.UnimplementedReservationServiceServer
	stripe      config.StripeConfig
	repo        repositories.ReservationRepository
	eventClient eventpb.EventServiceClient
}

func New(repo repositories.ReservationRepository, stripe config.StripeConfig, eventClient eventpb.EventServiceClient) *ReserveDomainImpl {
	return &ReserveDomainImpl{
		repo:        repo,
		stripe:      stripe,
		eventClient: eventClient,
	}
}
