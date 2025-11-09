package realtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/pkg/httpclient"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/google/uuid"
)

type Config struct {
	BaseURL string `env:"BASE_URL"`
}

type Service struct {
	config *Config
	client *httpclient.Client
}

func New(config *Config) *Service {
	client, err := httpclient.New(config.BaseURL)
	if err != nil {
		logger.Panic("can't initialze http client", err)
	}

	return &Service{
		config: config,
		client: client,
	}
}

type PushMessageRequest struct {
	UserID    uuid.UUID `json:"userId"`
	EventType string    `json:"eventType"`
	Data      any       `json:"data"`
}

func (s *Service) PushMessage(ctx context.Context, userID uuid.UUID, eventType string, data any) error {
	req := PushMessageRequest{
		UserID:    userID,
		EventType: eventType,
		Data:      data,
	}

	marshalReq, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request for get space campaign")
	}

	response, err := s.client.Post(ctx, "/", httpclient.RequestOptions{
		Body: marshalReq,
	})
	if err != nil {
		return errors.Wrap(err, "failed to push message")
	}

	if response.StatusCode() != 204 {
		return errors.New("failed to push message, non-204 response")
	}

	slogAttr := []any{
		slog.String("package", "realtime"),
		slog.String("action", "PushMessage"),
	}
	logger.InfoContext(ctx, "pushed message to realtime service successfully", slogAttr...)

	return nil
}
