package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/cp-rektmart/aconcert-microservice/event/db/codegen"
	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ListEvents Lists events with pagination
func (s *EventService) ListEvents(ctx context.Context, req *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Configurable offset
	const DEFAULT_LIMIT = 10
	const DEFAULT_PAGE = 0

	// Handle optional fields (proto3 optional fields are pointers in Go)
	limit := DEFAULT_LIMIT
	if req.Limit != nil {
		limit = int(*req.Limit)
	}

	page := DEFAULT_PAGE
	if req.Page != nil {
		page = int(*req.Page)
	}

	query := ""
	if req.Query != nil {
		query = *req.Query
	}

	events, err := s.queries.ListEvents(ctx, db.ListEventsParams{
		Limit:  int32(limit),
		Offset: int32(page * limit),
		Query:  fmt.Sprintf("%%%s%%", query),
	})
	if err != nil {
		return nil, errors.New("failed to list events")
	}

	var eventList []*eventpb.Event
	for _, event := range events {
		eventeventproto := &eventpb.Event{
			Id:          event.ID.String(),
			CreatedAt:   event.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:   event.UpdatedAt.Time.Format(time.RFC3339),
			Name:        event.Name,
			Description: event.Description.String,
			LocationId:  event.LocationID.String(),
			Artist:      event.Artist,
			EventDate:   event.EventDate.Time.Format(time.RFC3339),
			Thumbnail:   event.Thumbnail.String,
			Images:      event.Images,
		}
		eventList = append(eventList, eventeventproto)
	}

	return &eventpb.ListEventsResponse{Events: eventList}, nil
}

// GetEvent retrieves an event by ID
func (s *EventService) GetEvent(ctx context.Context, req *eventpb.GetEventRequest) (*eventpb.GetEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// validate uuid
	if _, err := uuid.Parse(req.Id); err != nil {
		return nil, errors.New("invalid UUID format")
	}

	event, err := s.queries.GetEventByID(ctx, pgtype.UUID{Bytes: uuid.MustParse(req.Id), Valid: true})
	if err != nil {
		return nil, errors.New("event not found")
	}

	eventeventproto := &eventpb.Event{
		Id:          event.ID.String(),
		CreatedAt:   event.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   event.UpdatedAt.Time.Format(time.RFC3339),
		Name:        event.Name,
		Description: event.Description.String,
		LocationId:  event.LocationID.String(),
		Artist:      event.Artist,
		EventDate:   event.EventDate.Time.Format(time.RFC3339),
		Thumbnail:   event.Thumbnail.String,
		Images:      event.Images,
	}

	return &eventpb.GetEventResponse{Event: eventeventproto}, nil
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(ctx context.Context, req *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newUUID := uuid.New()

	// Parse locationId from request
	locationUUID, err := uuid.Parse(req.GetLocationId())
	if err != nil {
		return nil, errors.New("invalid locationId format")
	}

	// Parse eventDate from request
	eventDate, err := time.Parse(time.RFC3339, req.GetEventDate())
	if err != nil {
		return nil, errors.New("invalid eventDate format")
	}

	id, err := s.queries.CreateEvent(ctx, db.CreateEventParams{
		ID:          pgtype.UUID{Bytes: newUUID, Valid: true},
		Name:        req.GetName(),
		Description: pgtype.Text{String: req.GetDescription(), Valid: true},
		LocationID:  pgtype.UUID{Bytes: locationUUID, Valid: true},
		Artist:      req.GetArtist(),
		EventDate:   pgtype.Timestamptz{Time: eventDate, Valid: true},
		Thumbnail:   pgtype.Text{String: req.GetThumbnail(), Valid: true},
		Images:      req.GetImages(),
	})
	if err != nil {
		return nil, errors.New("failed to create event")
	}

	return &eventpb.CreateEventResponse{Id: id.String()}, nil
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(ctx context.Context, req *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("UpdateEvent called with ID:", req.Id)

	parsedUUID, err := parsedUUID(req.Id)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}

	eventData, err := s.queries.GetEventByID(ctx, parsedUUID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Prepare update params, using existing values if optional fields are not set
	name := eventData.Name
	if req.Name != nil {
		name = req.GetName()
	}

	description := eventData.Description.String
	if req.Description != nil {
		description = req.GetDescription()
	}

	locationID := eventData.LocationID
	if req.LocationId != nil && *req.LocationId != "" {
		parsedLoc, err := uuid.Parse(req.GetLocationId())
		if err == nil {
			locationID = pgtype.UUID{Bytes: parsedLoc, Valid: true}
		}
	}

	artist := eventData.Artist
	if req.Artist != nil {
		artist = req.Artist
	}

	eventDate := eventData.EventDate
	if req.EventDate != nil && *req.EventDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, req.GetEventDate())
		if err == nil {
			eventDate = pgtype.Timestamptz{Time: parsedDate, Valid: true}
		}
	}

	thumbnail := eventData.Thumbnail.String
	if req.Thumbnail != nil {
		thumbnail = req.GetThumbnail()
	}

	images := eventData.Images
	if req.Images != nil {
		images = req.Images
	}

	updateParams := db.UpdateEventParams{
		ID:          parsedUUID,
		Name:        name,
		Description: pgtype.Text{String: description, Valid: true},
		LocationID:  locationID,
		Artist:      artist,
		EventDate:   eventDate,
		Thumbnail:   pgtype.Text{String: thumbnail, Valid: true},
		Images:      images,
	}

	_, err = s.queries.UpdateEvent(ctx, updateParams)
	if err != nil {
		return nil, errors.New("failed to update event")
	}

	return &eventpb.UpdateEventResponse{Id: parsedUUID.String()}, nil
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(ctx context.Context, req *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pgUUID, err := parsedUUID(req.Id)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}

	_, err = s.queries.DeleteEvent(ctx, pgUUID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	return &eventpb.DeleteEventResponse{Id: pgUUID.String()}, nil
}

func parsedUUID(id string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var pgUUID pgtype.UUID
	pgUUID.Bytes = parsed
	pgUUID.Valid = true
	return pgUUID, nil
}
