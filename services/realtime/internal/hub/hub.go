package hub

import (
	"context"
	"time"

	"slices"

	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/entities"
	"github.com/google/uuid"
)

// Message is what each client will receive over SSE.
type Message struct {
	EventType entities.EventType `json:"eventType"`
	Data      string             `json:"data"`
}

// Client is the per-connection channel.
type Client chan Message

type subscription struct {
	userID uuid.UUID
	client Client
}

type broadcast struct {
	userID uuid.UUID
	msg    Message
}

// Hub manages all active clients & broadcasts.
type Hub struct {
	clients    map[uuid.UUID][]Client
	register   chan subscription
	unregister chan subscription
	broadcast  chan broadcast
}

func New() *Hub {
	h := &Hub{
		clients:    make(map[uuid.UUID][]Client),
		register:   make(chan subscription),
		unregister: make(chan subscription),
		broadcast:  make(chan broadcast),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case sub := <-h.register:
			h.clients[sub.userID] = append(h.clients[sub.userID], sub.client)

		case sub := <-h.unregister:
			conns := h.clients[sub.userID]
			for i, c := range conns {
				if c == sub.client {
					close(c)
					h.clients[sub.userID] = slices.Delete(conns, i, i+1)
					break
				}
			}
			if len(h.clients[sub.userID]) == 0 {
				delete(h.clients, sub.userID)
			}

		case bc := <-h.broadcast:
			for _, c := range h.clients[bc.userID] {
				select {
				case c <- bc.msg:
				case <-time.After(time.Second):
					// drop if client can't keep up
				}
			}
		}
	}
}

// Register adds a new subscriber.
func (h *Hub) Register(ctx context.Context, userID uuid.UUID, client Client) {
	h.register <- subscription{userID, client}
}

// Unregister removes a subscriber.
func (h *Hub) Unregister(ctx context.Context, userID uuid.UUID, client Client) {
	h.unregister <- subscription{userID, client}
}

// Broadcast sends an event to all userID’s clients.
func (h *Hub) Broadcast(ctx context.Context, userID uuid.UUID, eventType entities.EventType, data string) {
	h.broadcast <- broadcast{
		userID: userID,
		msg: Message{
			EventType: eventType,
			Data:      data,
		},
	}
}
