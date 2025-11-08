package server

import (
	"github.com/cp-rektmart/aconcert-microservice/location/internal/repository"
	locationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/location"
	"go.mongodb.org/mongo-driver/mongo"
)

type LocationService struct {
	locationpb.UnimplementedLocationServiceServer
	locationRepo *repository.LocationRepository
}

func NewLocationService(db *mongo.Database) *LocationService {
	locationRepo := repository.NewLocationRepository(db, "locations")

	return &LocationService{
		locationRepo: locationRepo,
	}
}
