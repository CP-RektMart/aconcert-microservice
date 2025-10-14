package server

import (
	"context"

	"github.com/cockroachdb/errors/grpc/status"
	locationproto "github.com/cp-rektmart/aconcert-microservice/location/proto/location"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
)

func (s *LocationService) AddZone(ctx context.Context, req *locationproto.AddZoneRequest) (*locationproto.LocationIdResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location id: %v", err)
	}

	if err := s.locationRepo.AddZone(ctx, id, req.Zone); err != nil {
		return nil, err
	}

	return &locationproto.LocationIdResponse{Id: req.LocationId}, nil
}

func (s *LocationService) RemoveZone(ctx context.Context, req *locationproto.RemoveZoneRequest) (*locationproto.LocationIdResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location id: %v", err)
	}

	if err := s.locationRepo.RemoveZone(ctx, id, req.ZoneNumber); err != nil {
		return nil, err
	}

	return &locationproto.LocationIdResponse{Id: req.LocationId}, nil
}