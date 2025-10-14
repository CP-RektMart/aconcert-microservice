package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/entities"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	EventKey     = "realtime:event:"
	UserEventKey = "realtime:user_event:"
	EventTTL     = 3 * time.Minute
	UserEventTTL = 10 * time.Minute
)

type Repository interface {
	GetEvent(ctx context.Context, eventID uuid.UUID) (entities.EventData, error)
	SetEvent(ctx context.Context, event entities.EventData) error
	RemoveEvent(ctx context.Context, eventID uuid.UUID) error
	GetUserEvents(ctx context.Context, userID uuid.UUID) ([]entities.EventData, error)
	AddUserEvent(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) error
	RemoveUserEvent(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) error
}
type RepositoryImpl struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) Repository {
	return &RepositoryImpl{
		redisClient: redisClient,
	}
}

func (r *RepositoryImpl) formatEventKey(eventID uuid.UUID) string {
	return fmt.Sprintf("%s%s", EventKey, eventID.String())
}

func (r *RepositoryImpl) formatUserEventKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s", UserEventKey, userID.String())
}

func (r *RepositoryImpl) GetEvent(ctx context.Context, eventID uuid.UUID) (entities.EventData, error) {
	key := r.formatEventKey(eventID)
	data, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return entities.EventData{}, errors.Wrap(err, "failed to get event")
	}

	var event entities.EventData
	if err := json.Unmarshal(data, &event); err != nil {
		return entities.EventData{}, errors.Wrap(err, "failed to unmarshal event")
	}

	return event, nil
}

func (r *RepositoryImpl) SetEvent(ctx context.Context, event entities.EventData) error {
	key := r.formatEventKey(event.EventID)
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}

	if err := r.redisClient.Set(ctx, key, data, EventTTL).Err(); err != nil {
		return errors.Wrap(err, "failed to set event")
	}

	return nil
}

func (r *RepositoryImpl) RemoveEvent(ctx context.Context, eventID uuid.UUID) error {
	key := r.formatEventKey(eventID)
	if err := r.redisClient.Del(ctx, key).Err(); err != nil {
		return errors.Wrap(err, "failed to remove event")
	}

	return nil
}

func (r *RepositoryImpl) GetUserEvents(ctx context.Context, userID uuid.UUID) ([]entities.EventData, error) {
	key := r.formatUserEventKey(userID)
	eventIDStrs, err := r.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user events")
	}

	var events []entities.EventData
	for _, idStr := range eventIDStrs {
		eventID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}

		event, err := r.GetEvent(ctx, eventID)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				r.redisClient.SRem(ctx, key, idStr)
				continue
			}
			return nil, errors.Wrap(err, "failed to get event")
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *RepositoryImpl) AddUserEvent(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) error {
	key := r.formatUserEventKey(userID)

	_, err := r.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if err := pipe.SAdd(ctx, key, eventID.String()).Err(); err != nil {
			return errors.Wrap(err, "failed to add user event")
		}
		if err := pipe.Expire(ctx, key, UserEventTTL).Err(); err != nil {
			return errors.Wrap(err, "failed to set user event expiration")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to add user event in transaction")
	}

	return nil
}

func (r *RepositoryImpl) RemoveUserEvent(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) error {
	key := r.formatUserEventKey(userID)

	_, err := r.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if err := pipe.SRem(ctx, key, eventID.String()).Err(); err != nil {
			return errors.Wrap(err, "failed to remove user event")
		}
		if err := pipe.Expire(ctx, key, UserEventTTL).Err(); err != nil {
			return errors.Wrap(err, "failed to set user event expiration")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to remove user event in transaction")
	}

	return nil
}
