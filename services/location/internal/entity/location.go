package entity

import (
	locationproto "github.com/cp-rektmart/aconcert-microservice/location/proto/location"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LocationEntity struct {
	ID            primitive.ObjectID      `bson:"_id,omitempty"`
	VenueName     string                  `bson:"venue_name"`
	City          string                  `bson:"city"`
	StateProvince string                  `bson:"state_province"`
	Country       string                  `bson:"country"`
	Latitude      float64                 `bson:"latitude"`
	Longitude     float64                 `bson:"longitude"`
	Zones         []*locationproto.Zone   `bson:"zones"`
}