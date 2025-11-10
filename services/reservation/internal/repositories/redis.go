package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// StartExpirationListener listens for Redis key expiration events
// and publishes seat-available updates when seat reservations expire
func (r *ReservationImpl) StartExpirationListener(ctx context.Context) {
	// Subscribe to keyspace notifications for expired keys
	// Pattern: __keyevent@0__:expired
	pubsub := r.redisClient.PSubscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	ch := pubsub.Channel()

	logger.InfoContext(ctx, "StartExpirationListener: Listening for expired Redis keys")
	logger.InfoContext(ctx, "StartExpirationListener: Subscribed to __keyevent@0__:expired")

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				logger.WarnContext(ctx, "StartExpirationListener: Received nil message")
				continue
			}
			// msg.Payload contains the expired key name
			// Example: "seat:event-123:1:5:10"
			logger.InfoContext(ctx, "StartExpirationListener: Received expiration event",
				"channel", msg.Channel,
				"pattern", msg.Pattern,
				"payload", msg.Payload)
			r.handleExpiredKey(ctx, msg.Payload)

		case <-ctx.Done():
			log.Println("StartExpirationListener: Stopped")
			return
		}
	}
}

// handleExpiredKey processes an expired seat reservation key
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
