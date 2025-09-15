package grpcserver

import (
	"sync"
	"time"

	pb "github.com/cp-rektmart/aconcert-microservice/event/proto"
)

type EventServer struct {
	pb.UnimplementedEventServiceServer
	mu     sync.Mutex
	events map[string]*pb.Event
}

func NewEventServer() *EventServer {
	s := &EventServer{
		events: make(map[string]*pb.Event),
	}
	s.seedEvents()
	return s
}

func (s *EventServer) seedEvents() {
	now := time.Now().Format(time.RFC3339)

	s.events["1"] = &pb.Event{
		Id:        "1",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Rock Concert",
		Description: "An amazing rock concert.",
		LocationId: "loc-1",
		Artist:     []string{"Band A", "Band B"},
		EventDate:  time.Now().AddDate(0, 0, 7).Format(time.RFC3339), // 7 days later
		Thumbnail:  "rock_thumbnail.png",
		Images:     []string{"rock1.png", "rock2.png"},
	}

	s.events["2"] = &pb.Event{
		Id:        "2",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Jazz Night",
		Description: "Smooth jazz evening.",
		LocationId: "loc-2",
		Artist:     []string{"Jazz Trio"},
		EventDate:  time.Now().AddDate(0, 0, 14).Format(time.RFC3339), // 14 days later
		Thumbnail:  "jazz_thumbnail.png",
		Images:     []string{"jazz1.png", "jazz2.png"},
	}
}
