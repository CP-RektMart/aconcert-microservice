package service

import (
	"context"

	"github.com/cockroachdb/errors"
	db "github.com/cp-rektmart/aconcert-microservice/event/db/codegen"
	"github.com/cp-rektmart/aconcert-microservice/event/internal/utils"
	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
	"github.com/google/uuid"
)

func (s *EventService) GetEventZoneByEventId(ctx context.Context, req *eventpb.GetEventZoneByEventIdRequest) (*eventpb.GetEventZoneByEventIdResponse, error) {
	eventZones, err := s.queries.GetEventZonesByEventID(ctx, utils.ParsedUUID(req.EventId))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get event zones by event ID")
	}

	var eventZoneList []*eventpb.EventZone
	for _, zone := range eventZones {
		eventZoneList = append(eventZoneList, &eventpb.EventZone{
			Id:          zone.ID.String(),
			EventId:     zone.EventID.String(),
			LocationId:  zone.LocationID,
			ZoneNumber:  zone.ZoneNumber,
			Price:       zone.Price,
			Color:       zone.Color,
			Name:        zone.Name,
			Description: zone.Description,
			IsSoldOut:   zone.IsSoldOut,
		})
	}

	return &eventpb.GetEventZoneByEventIdResponse{List: eventZoneList}, nil
}

func (s *EventService) CreateEventZone(ctx context.Context, req *eventpb.CreateEventZoneRequest) (*eventpb.CreateEventZoneResponse, error) {
	id, err := s.queries.CreateEventZone(ctx, db.CreateEventZoneParams{
		EventID:     utils.ParsedUUID(req.EventId),
		LocationID:  req.LocationId,
		ZoneNumber:  req.ZoneNumber,
		Price:       req.Price,
		Color:       req.Color,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create event zone")
	}

	return &eventpb.CreateEventZoneResponse{
		Id: uuid.UUID(id.Bytes).String(),
	}, nil
}

func (s *EventService) UpdateEventZone(ctx context.Context, req *eventpb.UpdateEventZoneRequest) (*eventpb.UpdateEventZoneResponse, error) {
	eventZone, err := s.queries.GetEventZoneByID(ctx, utils.ParsedUUID(req.Id))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get event zone by ID")
	}

	params := db.UpdateEventZoneParams{
		ID: utils.ParsedUUID(req.Id),
	}

	if req.EventId != nil {
		params.EventID = utils.ParsedUUID(*req.EventId)
	} else {
		params.EventID = eventZone.EventID
	}

	if req.LocationId != nil {
		params.LocationID = *req.LocationId
	} else {
		params.LocationID = eventZone.LocationID
	}

	if req.ZoneNumber != nil {
		params.ZoneNumber = *req.ZoneNumber
	} else {
		params.ZoneNumber = eventZone.ZoneNumber
	}

	if req.Price != nil {
		params.Price = *req.Price
	} else {
		params.Price = eventZone.Price
	}

	if req.Color != nil {
		params.Color = *req.Color
	} else {
		params.Color = eventZone.Color
	}

	if req.Name != nil {
		params.Name = *req.Name
	} else {
		params.Name = eventZone.Name
	}

	if req.Description != nil {
		params.Description = *req.Description
	} else {
		params.Description = eventZone.Description
	}

	if req.IsSoldOut != nil {
		params.IsSoldOut = *req.IsSoldOut
	} else {
		params.IsSoldOut = eventZone.IsSoldOut
	}

	_, err = s.queries.UpdateEventZone(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update event zone")
	}

	return &eventpb.UpdateEventZoneResponse{
		Id: req.Id,
	}, nil
}

func (s *EventService) DeleteEventZone(ctx context.Context, req *eventpb.DeleteEventZoneRequest) (*eventpb.Empty, error) {
	_, err := s.queries.DeleteEventZone(ctx, utils.ParsedUUID(req.Id))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete event zone")
	}

	return &eventpb.Empty{}, nil
}
