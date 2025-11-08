package event

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
)

type EventService struct {
	client eventpb.EventServiceClient
}

func NewService(client eventpb.EventServiceClient) *EventService {
	return &EventService{
		client: client,
	}
}

func (s *EventService) TransformEventResponse(event *eventpb.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:          event.Id,
		Name:        event.Name,
		Description: event.Description,
		LocationID:  event.LocationId,
		Artist:      event.Artist,
		EventDate:   event.EventDate,
		Thumbnail:   event.Thumbnail,
		Images:      event.Images,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}
}

func (s *EventService) TransformEventZoneResponse(ez *eventpb.EventZone) dto.EventZoneResponse {
	return dto.EventZoneResponse{
		ID:          ez.Id,
		EventID:     ez.EventId,
		LocationID:  ez.LocationId,
		ZoneNumber:  int(ez.ZoneNumber),
		Price:       ez.Price,
		Color:       ez.Color,
		Name:        ez.Name,
		Description: ez.Description,
		IsSoldOut:   ez.IsSoldOut,
	}
}

func (s *EventService) ListEvents(ctx context.Context, req *dto.ListEventsRequest) ([]dto.EventResponse, error) {
	page := int32(req.Page)
	limit := int32(req.Limit)

	response, err := s.client.ListEvents(ctx, &eventpb.ListEventsRequest{
		Query:  &req.Query,
		SortBy: &req.SortBy,
		Order:  &req.Order,
		Page:   &page,
		Limit:  &limit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list events")
	}

	events := make([]dto.EventResponse, 0, len(response.Events))
	for _, event := range response.Events {
		events = append(events, s.TransformEventResponse(event))
	}

	return events, nil
}

func (s *EventService) GetEvent(ctx context.Context, req *dto.GetEventRequest) (dto.EventResponse, error) {
	response, err := s.client.GetEvent(ctx, &eventpb.GetEventRequest{
		Id: req.ID,
	})
	if err != nil {
		return dto.EventResponse{}, errors.Wrap(err, "failed to get event")
	}

	return s.TransformEventResponse(response.Event), nil
}

func (s *EventService) CreateEvent(ctx context.Context, req *dto.CreateEventRequest) (string, error) {
	response, err := s.client.CreateEvent(ctx, &eventpb.CreateEventRequest{
		Name:        req.Name,
		Description: req.Description,
		LocationId:  req.LocationID,
		Artist:      req.Artist,
		EventDate:   req.EventDate,
		Thumbnail:   req.Thumbnail,
		Images:      req.Images,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create event")
	}

	return response.Id, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, req *dto.UpdateEventRequest) (string, error) {
	response, err := s.client.UpdateEvent(ctx, &eventpb.UpdateEventRequest{
		Id:          req.ID,
		Name:        &req.Name,
		Description: &req.Description,
		LocationId:  &req.LocationID,
		Artist:      req.Artist,
		EventDate:   &req.EventDate,
		Thumbnail:   &req.Thumbnail,
		Images:      req.Images,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to update event")
	}

	return response.Id, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, req *dto.DeleteEventRequest) error {
	_, err := s.client.DeleteEvent(ctx, &eventpb.DeleteEventRequest{
		Id: req.ID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete event")
	}

	return nil
}

func (s *EventService) GetEventZonesByEventID(ctx context.Context, req *dto.GetEventZoneByEventIDRequest) ([]dto.EventZoneResponse, error) {
	response, err := s.client.GetEventZoneByEventId(ctx, &eventpb.GetEventZoneByEventIdRequest{
		EventId: req.EventID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list event zones")
	}

	eventZones := make([]dto.EventZoneResponse, 0, len(response.List))
	for _, ez := range response.List {
		eventZones = append(eventZones, s.TransformEventZoneResponse(ez))
	}

	return eventZones, nil
}

func (s *EventService) CreateEventZone(ctx context.Context, req *dto.CreateEventZoneRequest) (string, error) {
	response, err := s.client.CreateEventZone(ctx, &eventpb.CreateEventZoneRequest{
		EventId:     req.EventID,
		LocationId:  req.LocationID,
		ZoneNumber:  int32(req.ZoneNumber),
		Price:       req.Price,
		Color:       req.Color,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create event zone")
	}

	return response.Id, nil
}

func (s *EventService) UpdateEventZone(ctx context.Context, req *dto.UpdateEventZoneRequest) (string, error) {
	zoneNum := int32(req.ZoneNumber)

	response, err := s.client.UpdateEventZone(ctx, &eventpb.UpdateEventZoneRequest{
		Id:          req.ID,
		EventId:     &req.EventID,
		LocationId:  &req.LocationID,
		ZoneNumber:  &zoneNum,
		Price:       &req.Price,
		Color:       &req.Color,
		Name:        &req.Name,
		Description: &req.Description,
		IsSoldOut:   &req.IsSoldOut,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to update event zone")
	}

	return response.Id, nil
}

func (s *EventService) DeleteEventZone(ctx context.Context, req *dto.DeleteEventZoneRequest) error {
	_, err := s.client.DeleteEventZone(ctx, &eventpb.DeleteEventZoneRequest{
		Id: req.ID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete event zone")
	}

	return nil
}
