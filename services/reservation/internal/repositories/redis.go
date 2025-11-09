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

func (r *ReservationImpl) publishSeatUpdate(ctx context.Context, eventID string, seat SeatInfo, status entities.SeatStatus) {
	channel := "seats:all" // Single channel for simplicity

	message := map[string]any{
		"eventId":    eventID,
		"zoneNumber": seat.ZoneNumber,
		"row":        seat.RowNumber,
		"column":     seat.ColNumber,
		"status":     status,
		"timestamp":  time.Now().Unix(),
	}

	data, _ := json.Marshal(message)

	go func() {
		if err := r.redisClient.Publish(context.Background(), channel, data).Err(); err != nil {
			logger.ErrorContext(ctx, "Error to publish message pub/sub")
		}
	}()
}

func (r *ReservationImpl) StartExpirationListener(ctx context.Context) error {
	// Subscribe to keyspace notifications for expired keys
	// Pattern: __keyevent@0__:expired
	pubsub := r.redisClient.PSubscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			// msg.Payload contains the expired key name
			// Example: "seat:event-123:1:5:10"
			r.handleExpiredKey(ctx, msg.Payload)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleExpiredKey processes an expired seat reservation key
func (r *ReservationImpl) handleExpiredKey(ctx context.Context, key string) {
	// Parse the key: "seat:eventID:zone:row:col"
	parts := strings.Split(key, ":")

	// Expected format: ["seat", "eventID", "zone", "row", "col"]
	if len(parts) != 5 || parts[0] != "seat" {
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

	// Publish that this seat is now available
	r.publishSeatUpdate(ctx, eventID, seat, "AVAILABLE")

	// Optional: Log for debugging
	fmt.Printf("Seat expired and released: Event=%s, Zone=%d, Row=%d, Col=%d\n",
		eventID, zoneNumber, rowNumber, colNumber)
}
