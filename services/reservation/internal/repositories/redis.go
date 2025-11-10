package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/entities"
)

func (r *ReservationImpl) CreateReservationTemp(ctx context.Context, userID, reservationID string, ttl time.Duration) error {
	key := fmt.Sprintf("reservation:temp:%s:%s", userID, reservationID)
	return r.redisClient.Set(ctx, key, reservationID, ttl).Err()
}

func (r *ReservationImpl) GetReservationTimeLeft(ctx context.Context, userID, reservationID string) (time.Duration, error) {
	key := fmt.Sprintf("reservation:temp:%s:%s", userID, reservationID)
	return r.redisClient.TTL(ctx, key).Result()
}

func (r *ReservationImpl) DeleteReservationTemp(ctx context.Context, userID, reservationID string) error {
	key := fmt.Sprintf("reservation:temp:%s:%s", userID, reservationID)
	return r.redisClient.Del(ctx, key).Err()
}

func (r *ReservationImpl) CheckSeatAvailable(ctx context.Context, eventID string, seat SeatInfo) (bool, error) {
	key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)
	exists, err := r.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 0, nil
}

func (r *ReservationImpl) SetSeatReserved(ctx context.Context, eventID string, seat SeatInfo, reservationID string, ttl time.Duration) error {
	key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)

	r.publishSeatUpdate(ctx, eventID, seat, entities.SeatReserved)
	//FYI: Cache the persisted seat
	return r.redisClient.Set(ctx, key, reservationID, 0).Err()
}

func (r *ReservationImpl) SetSeatTempReserved(ctx context.Context, eventID string, seat SeatInfo, reservationID string, ttl time.Duration) error {
	key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)

	// pub/sub
	r.publishSeatUpdate(ctx, eventID, seat, entities.SeatPending)

	//FYI: Cache the seat for 15 days
	return r.redisClient.Set(ctx, key, reservationID, ttl).Err()
}

func (r *ReservationImpl) DeleteSeatReservation(ctx context.Context, eventID string, seat SeatInfo) error {
	key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)

	r.publishSeatUpdate(ctx, eventID, seat, entities.SeatAvailable)

	return r.redisClient.Del(ctx, key).Err()
}

func (r *ReservationImpl) CacheReservationSeats(ctx context.Context, reservationID string, seats []SeatInfo, ttl time.Duration) error {
	key := fmt.Sprintf("reservation:seats:%s", reservationID)
	data, err := json.Marshal(seats)
	if err != nil {
		return err
	}
	return r.redisClient.Set(ctx, key, data, ttl).Err()
}

func (r *ReservationImpl) GetReservationSeats(ctx context.Context, reservationID string) ([]SeatInfo, error) {
	key := fmt.Sprintf("reservation:seats:%s", reservationID)
	data, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var seats []SeatInfo
	if err := json.Unmarshal([]byte(data), &seats); err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *ReservationImpl) DeleteReservationSeats(ctx context.Context, reservationID string) error {
	key := fmt.Sprintf("reservation:seats:%s", reservationID)
	return r.redisClient.Del(ctx, key).Err()
}

// GetAllEventSeats retrieves all reserved/pending seats for a specific event
// PENDING seats come from Redis (temporary, with TTL)
// RESERVED seats come from Tickets table in database (confirmed, permanent)
func (r *ReservationImpl) GetAllEventSeats(ctx context.Context, eventID string) ([]SeatStatusInfo, error) {
	var seats []SeatStatusInfo

	// 1. Get PENDING seats from Redis
	pattern := fmt.Sprintf("seat:%s:*", eventID)
	iter := r.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Parse key format: "seat:eventID:zone:row:col"
		parts := strings.Split(key, ":")

		if len(parts) != 5 {
			continue // Skip malformed keys
		}

		var zoneNum, rowNum, colNum int32
		fmt.Sscanf(parts[2], "%d", &zoneNum)
		fmt.Sscanf(parts[3], "%d", &rowNum)
		fmt.Sscanf(parts[4], "%d", &colNum)

		// Check TTL - only include PENDING seats (TTL > 0)
		// RESERVED seats are in database, not Redis
		ttl, err := r.redisClient.TTL(ctx, key).Result()
		if err != nil {
			logger.ErrorContext(ctx, "Failed to get TTL for seat", "key", key)
			continue
		}

		// Only include if TTL > 0 (PENDING)
		// Skip if TTL = -1 (would be RESERVED, but should be in DB instead)
		if ttl > 0 {
			seats = append(seats, SeatStatusInfo{
				ZoneNumber: zoneNum,
				RowNumber:  rowNum,
				ColNumber:  colNum,
				Status:     string(entities.SeatPending),
			})
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	// 2. Get RESERVED seats from database (Tickets table)
	tickets, err := r.db.ListTicketsByEventID(ctx, stringToUUID(eventID))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get tickets from database", "error", err, "eventID", eventID)
		// Don't fail completely - return Redis seats even if DB query fails
		return seats, nil
	}

	// Add tickets as RESERVED seats
	for _, ticket := range tickets {
		seats = append(seats, SeatStatusInfo{
			ZoneNumber: ticket.ZoneNumber,
			RowNumber:  ticket.RowNumber,
			ColNumber:  ticket.ColNumber,
			Status:     string(entities.SeatReserved),
		})
	}

	logger.InfoContext(ctx, "Retrieved event seats",
		"eventID", eventID,
		"total", len(seats),
		"pending", len(seats)-len(tickets),
		"reserved", len(tickets))

	return seats, nil
}

// publishSeatUpdate broadcasts seat status changes via Redis Pub/Sub
func (r *ReservationImpl) publishSeatUpdate(ctx context.Context, eventID string, seat SeatInfo, status entities.SeatStatus) {
	channel := "seats:all" // Single channel for all events

	message := map[string]any{
		"eventId":    eventID,
		"zoneNumber": seat.ZoneNumber,
		"row":        seat.RowNumber,
		"column":     seat.ColNumber,
		"status":     string(status),
		"timestamp":  time.Now().Unix(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to marshal seat update message", "error", err)
		return
	}

	logger.InfoContext(ctx, "Publishing seat update to Redis",
		"channel", channel,
		"eventID", eventID,
		"zone", seat.ZoneNumber,
		"row", seat.RowNumber,
		"column", seat.ColNumber,
		"status", string(status),
		"message", string(data))

	// Publish in background (non-blocking)
	go func() {
		result := r.redisClient.Publish(context.Background(), channel, data)
		if err := result.Err(); err != nil {
			logger.ErrorContext(context.Background(), "Failed to publish seat update to Redis",
				"error", err,
				"channel", channel,
				"message", string(data))
		} else {
			logger.InfoContext(context.Background(), "Successfully published seat update",
				"channel", channel,
				"subscribers", result.Val(),
				"message", string(data))
		}
	}()
}

// publishSeatUpdatesBatch broadcasts multiple seat status changes in a single message
func (r *ReservationImpl) publishSeatUpdatesBatch(ctx context.Context, eventID string, seats []SeatInfo, status entities.SeatStatus) {
	if len(seats) == 0 {
		return
	}

	channel := "seats:all"
	timestamp := time.Now().Unix()

	// Create array of seat updates
	updates := make([]map[string]any, len(seats))
	for i, seat := range seats {
		updates[i] = map[string]any{
			"eventId":    eventID,
			"zoneNumber": seat.ZoneNumber,
			"row":        seat.RowNumber,
			"column":     seat.ColNumber,
			"status":     string(status),
			"timestamp":  timestamp,
		}
	}

	// Wrap in batch message
	batchMessage := map[string]any{
		"type":    "batch",
		"updates": updates,
	}

	data, err := json.Marshal(batchMessage)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to marshal batch seat update message", "error", err)
		return
	}

	logger.InfoContext(ctx, "Publishing BATCH seat update to Redis",
		"channel", channel,
		"eventID", eventID,
		"seats_count", len(seats),
		"status", string(status),
		"message_size", len(data))

	// Publish in background (non-blocking)
	go func() {
		result := r.redisClient.Publish(context.Background(), channel, data)
		if err := result.Err(); err != nil {
			logger.ErrorContext(context.Background(), "Failed to publish batch seat update to Redis",
				"error", err,
				"channel", channel,
				"seats_count", len(seats))
		} else {
			logger.InfoContext(context.Background(), "Successfully published BATCH seat update",
				"channel", channel,
				"subscribers", result.Val(),
				"seats_count", len(seats))
		}
	}()
}

// StartExpirationListener listens for Redis key expiration events
// and publishes seat-available updates when seat reservations expire.
// It batches multiple expired keys together to reduce pub/sub message count.
func (r *ReservationImpl) StartExpirationListener(ctx context.Context) {
	// Subscribe to keyspace notifications for expired keys
	// Pattern: __keyevent@0__:expired
	pubsub := r.redisClient.PSubscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	ch := pubsub.Channel()

	logger.InfoContext(ctx, "StartExpirationListener: Listening for expired Redis keys (BATCH MODE)")
	logger.InfoContext(ctx, "StartExpirationListener: Subscribed to __keyevent@0__:expired")

	// Buffer for batching expired keys
	pendingKeys := make([]string, 0, 100)
	batchTimer := time.NewTimer(50 * time.Millisecond)
	batchTimer.Stop() // Stop initially, start when first key arrives

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				logger.WarnContext(ctx, "StartExpirationListener: Received nil message")
				continue
			}

			// Add expired key to batch
			pendingKeys = append(pendingKeys, msg.Payload)

			logger.InfoContext(ctx, "StartExpirationListener: Key expired (buffering for batch)",
				"key", msg.Payload,
				"buffer_size", len(pendingKeys))

			// Start/reset batch timer
			// If this is the first key, start the timer
			// If timer already running, reset it to wait for more keys
			if !batchTimer.Stop() {
				select {
				case <-batchTimer.C:
				default:
				}
			}
			batchTimer.Reset(50 * time.Millisecond)

		case <-batchTimer.C:
			// Timer expired - process the batch
			if len(pendingKeys) > 0 {
				logger.InfoContext(ctx, "StartExpirationListener: Batch timer fired, processing batch",
					"batch_size", len(pendingKeys))

				r.handleExpiredKeysBatch(ctx, pendingKeys)

				// Reset buffer
				pendingKeys = pendingKeys[:0]
			}

		case <-ctx.Done():
			logger.InfoContext(ctx, "StartExpirationListener: Stopped")
			return
		}
	}
}

// handleExpiredKeysBatch processes multiple expired keys in a batch
func (r *ReservationImpl) handleExpiredKeysBatch(ctx context.Context, keys []string) {
	if len(keys) == 0 {
		return
	}

	logger.InfoContext(ctx, "handleExpiredKeysBatch: Processing batch of expired keys",
		"count", len(keys))

	// Group seats by eventID
	seatsByEvent := make(map[string][]SeatInfo)

	for _, key := range keys {
		// Parse the key: "seat:eventID:zone:row:col"
		parts := strings.Split(key, ":")

		// Expected format: ["seat", "eventID", "zone", "row", "col"]
		if len(parts) != 5 || parts[0] != "seat" {
			logger.WarnContext(ctx, "handleExpiredKeysBatch: Invalid key format, skipping",
				"key", key)
			continue
		}

		eventID := parts[1]

		// Parse seat coordinates
		var zoneNumber, rowNumber, colNumber int32
		fmt.Sscanf(parts[2], "%d", &zoneNumber)
		fmt.Sscanf(parts[3], "%d", &rowNumber)
		fmt.Sscanf(parts[4], "%d", &colNumber)

		seat := SeatInfo{
			ZoneNumber: zoneNumber,
			RowNumber:  rowNumber,
			ColNumber:  colNumber,
		}

		seatsByEvent[eventID] = append(seatsByEvent[eventID], seat)
	}

	// Publish batch update for each event
	for eventID, seats := range seatsByEvent {
		logger.InfoContext(ctx, "handleExpiredKeysBatch: Publishing batch for event",
			"eventID", eventID,
			"seats_count", len(seats))

		r.publishSeatUpdatesBatch(ctx, eventID, seats, entities.SeatAvailable)

		logger.InfoContext(ctx, "Seats batch expired and released",
			"eventID", eventID,
			"seats_count", len(seats))
	}
}

// handleExpiredKey processes an expired seat reservation key (DEPRECATED - keeping for backwards compatibility)
// This is now replaced by handleExpiredKeysBatch for better performance
func (r *ReservationImpl) handleExpiredKey(ctx context.Context, key string) {
	logger.InfoContext(ctx, "handleExpiredKey: Processing expired key", "key", key)

	// Parse the key: "seat:eventID:zone:row:col"
	parts := strings.Split(key, ":")

	logger.InfoContext(ctx, "handleExpiredKey: Key parts", "parts", parts, "length", len(parts))

	// Expected format: ["seat", "eventID", "zone", "row", "col"]
	if len(parts) != 5 || parts[0] != "seat" {
		logger.WarnContext(ctx, "handleExpiredKey: Invalid key format, ignoring",
			"key", key,
			"parts_count", len(parts),
			"first_part", parts[0])
		return // Not a seat key, ignore
	}

	eventID := parts[1]

	// Parse seat coordinates
	var zoneNumber, rowNumber, colNumber int32
	fmt.Sscanf(parts[2], "%d", &zoneNumber)
	fmt.Sscanf(parts[3], "%d", &rowNumber)
	fmt.Sscanf(parts[4], "%d", &colNumber)

	seat := SeatInfo{
		ZoneNumber: zoneNumber,
		RowNumber:  rowNumber,
		ColNumber:  colNumber,
	}

	logger.InfoContext(ctx, "handleExpiredKey: Publishing AVAILABLE status",
		"eventID", eventID,
		"zone", zoneNumber,
		"row", rowNumber,
		"column", colNumber)

	// Publish that this seat is now available
	r.publishSeatUpdate(ctx, eventID, seat, entities.SeatAvailable)

	logger.InfoContext(ctx, "Seat expired and released",
		"eventID", eventID,
		"zone", zoneNumber,
		"row", rowNumber,
		"column", colNumber)
}

// SetSeatsReservedBatch marks multiple seats as RESERVED in a single batch operation
// This is used when confirming a reservation with multiple seats
func (r *ReservationImpl) SetSeatsReservedBatch(ctx context.Context, eventID string, seats []SeatInfo, reservationID string) error {
	if len(seats) == 0 {
		return nil
	}

	logger.InfoContext(ctx, "SetSeatsReservedBatch: Marking seats as RESERVED",
		"eventID", eventID,
		"reservationID", reservationID,
		"seats_count", len(seats))

	// 1. Update all seats in Redis (set TTL=0 for permanent)
	for _, seat := range seats {
		key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)
		if err := r.redisClient.Set(ctx, key, reservationID, 0).Err(); err != nil {
			logger.ErrorContext(ctx, "Failed to set seat reserved in Redis",
				"error", err,
				"seat", seat)
			return err
		}
	}

	// 2. Publish batch update (single message for all seats)
	r.publishSeatUpdatesBatch(ctx, eventID, seats, entities.SeatReserved)

	logger.InfoContext(ctx, "SetSeatsReservedBatch: Successfully marked seats as RESERVED",
		"seats_count", len(seats))

	return nil
}

// SetSeatsTempReservedBatch marks multiple seats as PENDING in a single batch operation
// This is used when creating a reservation with multiple seats
func (r *ReservationImpl) SetSeatsTempReservedBatch(ctx context.Context, eventID string, seats []SeatInfo, reservationID string, ttl time.Duration) error {
	if len(seats) == 0 {
		return nil
	}

	logger.InfoContext(ctx, "SetSeatsTempReservedBatch: Marking seats as PENDING",
		"eventID", eventID,
		"reservationID", reservationID,
		"seats_count", len(seats),
		"ttl", ttl)

	// 1. Update all seats in Redis (set TTL for temporary hold)
	for _, seat := range seats {
		key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)
		if err := r.redisClient.Set(ctx, key, reservationID, ttl).Err(); err != nil {
			logger.ErrorContext(ctx, "Failed to set seat temp reserved in Redis",
				"error", err,
				"seat", seat)
			return err
		}
	}

	// 2. Publish batch update (single message for all seats)
	r.publishSeatUpdatesBatch(ctx, eventID, seats, entities.SeatPending)

	logger.InfoContext(ctx, "SetSeatsTempReservedBatch: Successfully marked seats as PENDING",
		"seats_count", len(seats))

	return nil
}
