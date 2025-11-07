package location

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	locationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/location"
)

type LocationService struct {
	client locationpb.LocationServiceClient
}

func NewService(client locationpb.LocationServiceClient) *LocationService {
	return &LocationService{
		client: client,
	}
}

func (s *LocationService) TransformZoneResponses(zone *locationpb.Zone) dto.ZoneResponse {
	return dto.ZoneResponse{
		ZoneNumber:   int(zone.ZoneNumber),
		ZoneName:     zone.ZoneName,
		Capacity:     int(zone.Capacity),
		SeatsPerRow:  int(zone.SeatsPerRow),
		NumberOfRows: int(zone.NumberOfRows),
	}
}

func (s *LocationService) TransformLocationResponse(location *locationpb.Location) dto.LocationResponse {
	zones := make([]dto.ZoneResponse, 0, len(location.Zones))
	for _, zone := range location.Zones {
		zones = append(zones, s.TransformZoneResponses(zone))
	}

	return dto.LocationResponse{
		ID:            location.Id,
		VenueName:     location.VenueName,
		City:          location.City,
		StateProvince: location.StateProvince,
		Country:       location.Country,
		Latitude:      location.Latitude,
		Longitude:     location.Longitude,
		Zones:         zones,
	}
}

func (s *LocationService) ListLocations(ctx context.Context, req *dto.ListLocationsRequest) (dto.ListLocationsResponse, error) {
	response, err := s.client.ListLocations(ctx, &locationpb.ListLocationsRequest{})
	if err != nil {
		return dto.ListLocationsResponse{}, errors.Wrap(err, "failed to list locations")
	}

	locations := make([]dto.LocationResponse, 0, len(response.Locations))
	for _, location := range response.Locations {
		locations = append(locations, s.TransformLocationResponse(location))
	}

	return dto.ListLocationsResponse{
		List: locations,
	}, nil
}

func (s *LocationService) GetLocation(ctx context.Context, req *dto.GetLocationRequest) (dto.LocationResponse, error) {
	response, err := s.client.GetLocation(ctx, &locationpb.GetLocationRequest{
		Id: req.ID,
	})
	if err != nil {
		return dto.LocationResponse{}, errors.Wrap(err, "failed to get location")
	}

	return s.TransformLocationResponse(response), nil
}

func (s *LocationService) CreateLocation(ctx context.Context, req *dto.CreateLocationRequest) (dto.CreateLocationResponse, error) {
	zones := make([]*locationpb.Zone, 0, len(req.Zones))
	for _, zoneReq := range req.Zones {
		zones = append(zones, &locationpb.Zone{
			ZoneNumber:   int32(zoneReq.ZoneNumber),
			ZoneName:     zoneReq.ZoneName,
			Capacity:     int32(zoneReq.Capacity),
			SeatsPerRow:  int32(zoneReq.SeatsPerRow),
			NumberOfRows: int32(zoneReq.NumberOfRows),
		})
	}

	response, err := s.client.CreateLocation(ctx, &locationpb.Location{
		VenueName:     req.VenueName,
		City:          req.City,
		StateProvince: req.StateProvince,
		Country:       req.Country,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Zones:         zones,
	})
	if err != nil {
		return dto.CreateLocationResponse{}, nil
	}

	return dto.CreateLocationResponse{
		ID: response.Id,
	}, nil
}

func (s *LocationService) UpdateLocation(ctx context.Context, req *dto.UpdateLocationRequest) (dto.UpdateLocationResponse, error) {
	zones := make([]*locationpb.Zone, 0, len(req.Zones))
	for _, zoneReq := range req.Zones {
		zones = append(zones, &locationpb.Zone{
			ZoneNumber:   int32(zoneReq.ZoneNumber),
			ZoneName:     zoneReq.ZoneName,
			Capacity:     int32(zoneReq.Capacity),
			SeatsPerRow:  int32(zoneReq.SeatsPerRow),
			NumberOfRows: int32(zoneReq.NumberOfRows),
		})
	}

	response, err := s.client.UpdateLocation(ctx, &locationpb.UpdateLocationRequest{
		Id:            req.ID,
		VenueName:     req.VenueName,
		City:          req.City,
		StateProvince: req.StateProvince,
		Country:       req.Country,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Zones:         zones,
	})
	if err != nil {
		return dto.UpdateLocationResponse{}, errors.Wrap(err, "failed to update location")
	}

	return dto.UpdateLocationResponse{
		ID: response.Id,
	}, nil
}

func (s *LocationService) DeleteLocation(ctx context.Context, req *dto.DeleteLocationRequest) error {
	_, err := s.client.DeleteLocation(ctx, &locationpb.DeleteLocationRequest{
		Id: req.ID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete location")
	}

	return nil
}
