package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/pkg/httpclient"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
)

type AuthService struct {
	client *httpclient.Client
}

func NewService(baseUrl string) *AuthService {
	client, err := httpclient.New(baseUrl)
	if err != nil {
		logger.Panic("can't initialze http client", err)
	}

	return &AuthService{
		client: client,
	}
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (dto.LoginResponse, error) {
	marshalPayload, err := json.Marshal(req)
	if err != nil {
		return dto.LoginResponse{}, errors.Wrap(err, "failed to marshal payload")
	}
	response, err := s.client.Post(ctx, "/v1/auth/login", httpclient.RequestOptions{
		Body: marshalPayload,
	})
	if err != nil {
		return dto.LoginResponse{}, errors.Wrap(err, "failed to enqueue task")
	}

	data := &dto.LoginResponse{}
	if err = json.Unmarshal(response.Body(), data); err != nil {
		return dto.LoginResponse{}, errors.Wrap(err, "failed to unmarshal get space campaigns response")
	}

	return *data, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error) {
	marshalPayload, err := json.Marshal(req)
	if err != nil {
		return dto.RefreshTokenResponse{}, errors.Wrap(err, "failed to marshal payload")
	}
	response, err := s.client.Post(ctx, "/v1/auth/refresh", httpclient.RequestOptions{
		Body: marshalPayload,
	})
	if err != nil {
		return dto.RefreshTokenResponse{}, errors.Wrap(err, "failed to enqueue task")
	}

	data := &dto.RefreshTokenResponse{}
	if err = json.Unmarshal(response.Body(), data); err != nil {
		return dto.RefreshTokenResponse{}, errors.Wrap(err, "failed to unmarshal get space campaigns response")
	}

	return *data, nil
}

func (s *AuthService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	marshalPayload, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal payload")
	}
	response, err := s.client.Post(ctx, "/v1/auth/logout", httpclient.RequestOptions{
		Body: marshalPayload,
	})
	if err != nil {
		return errors.Wrap(err, "failed to enqueue task")
	}

	if response.StatusCode() != http.StatusNoContent {
		return errors.New("unexpected status code")
	}

	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, req *dto.GetProfileRequest) (dto.UserResponse, error) {
	marshalPayload, err := json.Marshal(req)
	if err != nil {
		return dto.UserResponse{}, errors.Wrap(err, "failed to marshal payload")
	}
	response, err := s.client.Get(ctx, "/v1/auth/me", httpclient.RequestOptions{
		Body: marshalPayload,
	})
	if err != nil {
		return dto.UserResponse{}, errors.Wrap(err, "failed to enqueue task")
	}

	data := &dto.UserResponse{}
	if err = json.Unmarshal(response.Body(), data); err != nil {
		return dto.UserResponse{}, errors.Wrap(err, "failed to unmarshal get space campaigns response")
	}

	return *data, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (dto.UserResponse, error) {
	marshalPayload, err := json.Marshal(req)
	if err != nil {
		return dto.UserResponse{}, errors.Wrap(err, "failed to marshal payload")
	}
	response, err := s.client.Patch(ctx, "/v1/auth/me", httpclient.RequestOptions{
		Body: marshalPayload,
	})
	if err != nil {
		return dto.UserResponse{}, errors.Wrap(err, "failed to enqueue task")
	}

	data := &dto.UserResponse{}
	if err = json.Unmarshal(response.Body(), data); err != nil {
		return dto.UserResponse{}, errors.Wrap(err, "failed to unmarshal get space campaigns response")
	}

	return *data, nil
}
