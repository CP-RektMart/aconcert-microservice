package service

import (
	"sync"

	db "github.com/cp-rektmart/aconcert-microservice/event/db/codegen"
	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
)

type EventService struct {
	eventpb.UnimplementedEventServiceServer
	mu      sync.Mutex
	queries *db.Queries
}

func NewEventService(queries *db.Queries) *EventService {
	return &EventService{
		queries: queries,
	}
}
