package domain

type ReserveDomain interface{}

type ReserveDomainImpl struct{}

func New() ReserveDomain {
	return &ReserveDomainImpl{}
}