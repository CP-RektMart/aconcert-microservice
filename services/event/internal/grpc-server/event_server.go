package grpcserver

import (
	"sync"

	pb "github.com/cp-rektmart/aconcert-microservice/event/internal/proto"
	"github.com/cp-rektmart/aconcert-microservice/event/internal/store"
)

type EventServer struct {
	pb.UnimplementedEventServiceServer
	mu    sync.Mutex
	store *store.Store
}

func NewEventServer(store *store.Store) *EventServer {
	return &EventServer{
		store: store,
	}
}
