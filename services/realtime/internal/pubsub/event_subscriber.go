package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/hub"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// EventSubscriber manages Redis Pub/Sub subscriptions for seat updates
// Multiple users viewing the same event share ONE Redis subscription (connection pooling)
type EventSubscriber struct {
	redisClient *redis.Client
	hub         *hub.Hub
	pubsub      *redis.PubSub
	cancel      context.CancelFunc
	mu          sync.RWMutex
	eventUsers  map[string]map[uuid.UUID]bool // eventID -> set of userIDs
}

// NewEventSubscriber creates and starts a new event subscriber
func NewEventSubscriber(redisClient *redis.Client, hub *hub.Hub) *EventSubscriber {
	ctx, cancel := context.WithCancel(context.Background())

	es := &EventSubscriber{
		redisClient: redisClient,
		hub:         hub,
		cancel:      cancel,
		eventUsers:  make(map[string]map[uuid.UUID]bool),
	}

	// Subscribe to the global seat updates channel
	es.pubsub = redisClient.Subscribe(ctx, "seats:all")

	// Start message forwarding in background
	go es.forwardMessages(ctx)

	log.Println("EventSubscriber: Started listening to 'seats:all' channel")

	return es
}

// forwardMessages listens to Redis Pub/Sub and forwards to interested users
func (es *EventSubscriber) forwardMessages(ctx context.Context) {
	ch := es.pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			es.handleMessage(ctx, msg)

		case <-ctx.Done():
			log.Println("EventSubscriber: Stopped")
			return
		}
	}
}

// handleMessage processes a single pub/sub message (either single or batch)
func (es *EventSubscriber) handleMessage(ctx context.Context, msg *redis.Message) {
	// Try parsing as batch first
	var batch SeatUpdateBatch
	if err := json.Unmarshal([]byte(msg.Payload), &batch); err == nil && batch.Type == "batch" {
		// Handle batch message
		log.Printf("📦 EventSubscriber: Received BATCH message with %d updates", len(batch.Updates))

		if !batch.IsValid() {
			log.Printf("EventSubscriber: Invalid batch message: %+v", batch)
			return
		}

		es.handleBatchUpdate(ctx, batch)
		return
	}

	// Fallback to single update (backwards compatible)
	var seatUpdate SeatUpdate
	if err := json.Unmarshal([]byte(msg.Payload), &seatUpdate); err != nil {
		log.Printf("EventSubscriber: Failed to parse message: %v", err)
		return
	}

	// Validate single update
	if !seatUpdate.IsValid() {
		log.Printf("EventSubscriber: Invalid seat update: %+v", seatUpdate)
		return
	}

	// Process single update
	es.forwardSingleUpdate(ctx, seatUpdate, msg.Payload)
}

// handleBatchUpdate processes a batch of seat updates
func (es *EventSubscriber) handleBatchUpdate(ctx context.Context, batch SeatUpdateBatch) {
	if len(batch.Updates) == 0 {
		return
	}

	// Group updates by eventID
	updatesByEvent := make(map[string][]SeatUpdate)
	for _, update := range batch.Updates {
		updatesByEvent[update.EventID] = append(updatesByEvent[update.EventID], update)
	}

	// Forward to subscribers for each event
	totalForwarded := 0
	for eventID, updates := range updatesByEvent {
		es.mu.RLock()
		users, exists := es.eventUsers[eventID]
		if !exists || len(users) == 0 {
			es.mu.RUnlock()
			log.Printf("EventSubscriber: Batch updates for event '%s' dropped - no subscribers (%d updates)",
				eventID, len(updates))
			continue
		}

		// Copy user IDs
		userIDs := make([]uuid.UUID, 0, len(users))
		for userID := range users {
			userIDs = append(userIDs, userID)
		}
		es.mu.RUnlock()

		// Forward each update to all users
		log.Printf("📦 Processing batch update for event %s with %d seat updates for %d users",
			eventID, len(updates), len(userIDs))

		for i, update := range updates {
			// Marshal individual update
			data, err := json.Marshal(update)
			if err != nil {
				log.Printf("EventSubscriber: Failed to marshal update: %v", err)
				continue
			}

			// Check if users are still subscribed before broadcasting each update
			es.mu.RLock()
			currentUsers, stillExists := es.eventUsers[eventID]
			currentUserCount := 0
			if stillExists {
				currentUserCount = len(currentUsers)
			}
			es.mu.RUnlock()

			// Warn if user count changed during batch processing
			if currentUserCount != len(userIDs) {
				log.Printf("⚠️  User count changed during batch! Started with %d, now %d (update %d/%d for event %s, seat %d-%d-%d)",
					len(userIDs), currentUserCount, i+1, len(updates), eventID, update.ZoneNumber, update.Row, update.Column)
			}

			for _, userID := range userIDs {
				es.hub.Broadcast(ctx, userID, "SEAT", string(data))
			}
			totalForwarded++

			log.Printf("EventSubscriber: Forwarded to %d users - Event: %s, Seat: %d-%d-%d, Status: %s",
				len(userIDs), update.EventID, update.ZoneNumber, update.Row, update.Column, update.Status)
		}

		log.Printf("✅ EventSubscriber: Completed forwarding %d seat updates to %d users for event %s",
			len(updates), len(userIDs), eventID)
	}

	log.Printf("📦 EventSubscriber: Batch processing complete - %d total updates forwarded", totalForwarded)
}

// forwardSingleUpdate forwards a single seat update to interested users
func (es *EventSubscriber) forwardSingleUpdate(ctx context.Context, update SeatUpdate, payload string) {
	// Find users watching this event
	es.mu.RLock()
	users, exists := es.eventUsers[update.EventID]
	if !exists || len(users) == 0 {
		es.mu.RUnlock()
		log.Printf("⚠️  EventSubscriber: Single update for event '%s' dropped - no subscribers (Seat: %d-%d-%d, Status: %s)",
			update.EventID, update.ZoneNumber, update.Row, update.Column, update.Status)
		return
	}

	// Copy user IDs to avoid holding lock during broadcast
	userIDs := make([]uuid.UUID, 0, len(users))
	for userID := range users {
		userIDs = append(userIDs, userID)
	}
	es.mu.RUnlock()

	// Forward to all interested users
	for _, userID := range userIDs {
		es.hub.Broadcast(ctx, userID, "SEAT", payload)
	}

	log.Printf("EventSubscriber: Forwarded to %d users - Event: %s, Seat: %d-%d-%d, Status: %s",
		len(userIDs), update.EventID, update.ZoneNumber, update.Row, update.Column, update.Status)
}

// Subscribe registers a user's interest in an event
func (es *EventSubscriber) Subscribe(ctx context.Context, userID uuid.UUID, eventID string) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Initialize event map if needed
	if es.eventUsers[eventID] == nil {
		es.eventUsers[eventID] = make(map[uuid.UUID]bool)
		log.Printf("EventSubscriber: First subscriber for event %s", eventID)
	}

	// Add user to event's subscriber list
	es.eventUsers[eventID][userID] = true

	log.Printf("✨ EventSubscriber: User %s subscribed to event %s (%d total users)",
		userID, eventID, len(es.eventUsers[eventID]))

	return nil
}

// Unsubscribe removes a user from an event
func (es *EventSubscriber) Unsubscribe(ctx context.Context, userID uuid.UUID, eventID string) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	users, exists := es.eventUsers[eventID]
	if !exists {
		return nil
	}

	delete(users, userID)

	// Clean up empty event
	if len(users) == 0 {
		delete(es.eventUsers, eventID)
		log.Printf("👋 EventSubscriber: Last subscriber left event %s", eventID)
	} else {
		log.Printf("👋 EventSubscriber: User %s unsubscribed from event %s (%d remaining)",
			userID, eventID, len(users))
	}

	return nil
}

// UnsubscribeUserFromAll removes a user from all events
func (es *EventSubscriber) UnsubscribeUserFromAll(ctx context.Context, userID uuid.UUID) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	for eventID, users := range es.eventUsers {
		if users[userID] {
			delete(users, userID)

			if len(users) == 0 {
				delete(es.eventUsers, eventID)
			}
		}
	}

	return nil
}

// Close stops the subscriber and closes Redis connection
func (es *EventSubscriber) Close() error {
	log.Println("EventSubscriber: Shutting down...")

	es.cancel()

	if err := es.pubsub.Close(); err != nil {
		log.Printf("EventSubscriber: Error closing pubsub: %v", err)
		return err
	}

	log.Println("EventSubscriber: Shutdown complete")
	return nil
}
