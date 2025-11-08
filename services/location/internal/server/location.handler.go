package server

import (
	"context"

	"github.com/cockroachdb/errors/grpc/status"
	"github.com/cp-rektmart/aconcert-microservice/location/internal/entity"
	locationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/location"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
)

func (s *LocationService) CreateLocation(ctx context.Context, req *locationpb.Location) (*locationpb.LocationIdResponse, error) {
	if err := validateZones(req.Zones); err != nil {
		return nil, err
	}

	id, err := s.locationRepo.Insert(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "insert failed: %v", err)
	}

	return &locationpb.LocationIdResponse{Id: id.Hex()}, nil
}

func (s *LocationService) GetLocation(ctx context.Context, req *locationpb.GetLocationRequest) (*locationpb.Location, error) {
	objID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	loc, err := s.locationRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "not found: %v", err)
	}

	return toProtoLocation(loc), nil
}

func (s *LocationService) ListLocations(ctx context.Context, _ *locationpb.ListLocationsRequest) (*locationpb.ListLocationsResponse, error) {
	locs, err := s.locationRepo.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list: %v", err)
	}

	return toProtoList(locs), nil
}

func (s *LocationService) UpdateLocation(ctx context.Context, req *locationpb.UpdateLocationRequest) (*locationpb.LocationIdResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	updateFields := collectUpdateFields(req)
	if len(updateFields) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no fields to update")
	}

	if err := s.locationRepo.Update(ctx, objID, updateFields); err != nil {
		return nil, status.Errorf(codes.Internal, "update failed: %v", err)
	}

	return &locationpb.LocationIdResponse{Id: req.Id}, nil
}

func (s *LocationService) DeleteLocation(ctx context.Context, req *locationpb.DeleteLocationRequest) (*locationpb.DeleteLocationResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	success, err := s.locationRepo.Delete(ctx, objID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete failed: %v", err)
	}

	return &locationpb.DeleteLocationResponse{Success: success}, nil
}

func validateZones(zones []*locationpb.Zone) error {
	seen := map[int32]bool{}
	for _, z := range zones {
		if seen[z.ZoneNumber] {
			return status.Errorf(codes.InvalidArgument, "duplicate zone_number: %d", z.ZoneNumber)
		}
		seen[z.ZoneNumber] = true
	}
	return nil
}

func collectUpdateFields(req *locationpb.UpdateLocationRequest) bson.M {
	fields := bson.M{}

	if req.VenueName != "" {
		fields["venue_name"] = req.VenueName
	}
	if req.City != "" {
		fields["city"] = req.City
	}
	if req.StateProvince != "" {
		fields["state_province"] = req.StateProvince
	}
	if req.Country != "" {
		fields["country"] = req.Country
	}
	if req.Latitude != 0 {
		fields["latitude"] = req.Latitude
	}
	if req.Longitude != 0 {
		fields["longitude"] = req.Longitude
	}
	if len(req.Zones) > 0 {
		if err := validateZones(req.Zones); err != nil {
			return fields
		}
		fields["zones"] = req.Zones
	}

	return fields
}

func toProtoLocation(loc *entity.LocationEntity) *locationpb.Location {
	return &locationpb.Location{
		Id:            loc.ID.Hex(),
		VenueName:     loc.VenueName,
		City:          loc.City,
		StateProvince: loc.StateProvince,
		Country:       loc.Country,
		Latitude:      loc.Latitude,
		Longitude:     loc.Longitude,
		Zones:         loc.Zones,
	}
}

func toProtoList(locs []*entity.LocationEntity) *locationpb.ListLocationsResponse {
	protoLocs := make([]*locationpb.Location, len(locs))
	for i, l := range locs {
		protoLocs[i] = toProtoLocation(l)
	}
	return &locationpb.ListLocationsResponse{Locations: protoLocs}
}
