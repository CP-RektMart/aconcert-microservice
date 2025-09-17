package grpcserver

import (
	"sync"

	db "github.com/cp-rektmart/aconcert-microservice/event/db/codegen"
	eventproto "github.com/cp-rektmart/aconcert-microservice/event/proto/event"
)

type EventService struct {
	eventproto.UnimplementedEventServiceServer
	mu      sync.Mutex
	queries *db.Queries
}

func NewEventService(queries *db.Queries) *EventService {
	return &EventService{
		queries: queries,
	}
}
