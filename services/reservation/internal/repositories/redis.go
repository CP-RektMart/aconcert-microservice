package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
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
	//FYI: Cache the seat for 15 days
	return r.redisClient.Set(ctx, key, reservationID, ttl*24).Err()
}

func (r *ReservationImpl) SetSeatTempReserved(ctx context.Context, eventID string, seat SeatInfo, reservationID string, ttl time.Duration) error {
	key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)
	//FYI: Cache the seat for 15 days
	return r.redisClient.Set(ctx, key, reservationID, ttl).Err()
}

func (r *ReservationImpl) DeleteSeatReservation(ctx context.Context, eventID string, seat SeatInfo) error {
	key := fmt.Sprintf("seat:%s:%d:%d:%d", eventID, seat.ZoneNumber, seat.RowNumber, seat.ColNumber)
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
