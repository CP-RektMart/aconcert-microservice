package grpcserver

import (
	"context"
	"errors"
	"strings"
	"time"

	pb "github.com/cp-rektmart/aconcert-microservice/event/proto"
)

// ListEvents Lists events with pagination
func (s *EventServer) ListEvents(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []*pb.Event
	for _, e := range s.events {
		if req.Query == "" || containsIgnoreCase(e.Name, req.Query) || containsIgnoreCase(e.Description, req.Query) {
			filtered = append(filtered, e)
		}
	}

	total := int32(len(filtered))
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	pagination := &pb.Pagination{
		Total:      total,
		Count:      int32(end - start),
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	}

	return &pb.ListEventsResponse{
		Events:     filtered[start:end],
		Pagination: pagination,
	}, nil
}

// GetEvent retrieves an event by ID
func (s *EventServer) GetEvent(ctx context.Context, req *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[req.Id]
	if !ok {
		return nil, errors.New("event not found")
	}
	return &pb.GetEventResponse{Event: event}, nil
}

// CreateEvent creates a new event
func (s *EventServer) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := time.Now().Format("20060102150405") // unique id
	now := time.Now().Format(time.RFC3339)

	event := &pb.Event{
		Id:         id,
		CreatedAt:  now,
		UpdatedAt:  now,
		Name:       req.GetName(),
		Description: req.GetDescription(),
		LocationId: req.GetLocationId(),
		Artist:     req.GetArtist(),
		EventDate:  req.GetEventDate(), // expect string in RFC3339
		Thumbnail:  req.GetThumbnail(),
		Images:     req.GetImages(),
	}

	s.events[id] = event
	return &pb.CreateEventResponse{Id: event.Id}, nil
}

// UpdateEvent updates an existing event
func (s *EventServer) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[req.Id]
	if !ok {
		return nil, errors.New("event not found")
	}

	// Update fields only if provided
	if req.GetName() != "" {
		event.Name = req.GetName()
	}
	if req.GetDescription() != "" {
		event.Description = req.GetDescription()
	}
	if req.GetLocationId() != "" {
		event.LocationId = req.GetLocationId()
	}
	if len(req.GetArtist()) > 0 {
		event.Artist = req.GetArtist()
	}
	if req.GetEventDate() != "" {
		event.EventDate = req.GetEventDate()
	}
	if req.GetThumbnail() != "" {
		event.Thumbnail = req.GetThumbnail()
	}
	if len(req.GetImages()) > 0 {
		event.Images = req.GetImages()
	}

	// Always update the UpdatedAt timestamp
	event.UpdatedAt = time.Now().Format(time.RFC3339)

	return &pb.UpdateEventResponse{Id: event.Id}, nil
}

// DeleteEvent deletes an event
func (s *EventServer) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[req.Id]
	if !ok {
		return nil, errors.New("event not found")
	}

	event.DeletedAt = time.Now().Format(time.RFC3339)
	delete(s.events, req.Id)

	return &pb.DeleteEventResponse{}, nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
