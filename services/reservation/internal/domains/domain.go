package domains

import (
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/repositories"
)

type ReserveDomain interface {
	Reserve() error
	Confirm() error
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
