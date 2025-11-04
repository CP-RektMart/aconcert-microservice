package domains

import reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"

type ReserveDomain interface{
	Reserve() error
	Confirm() error
}

type ReserveDomainImpl struct {
	reservationpb.UnimplementedReservationServiceServer
	// cache
	// db
}

func New() *ReserveDomainImpl {
	return &ReserveDomainImpl{}
}