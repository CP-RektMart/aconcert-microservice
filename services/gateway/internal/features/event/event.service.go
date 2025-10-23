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
