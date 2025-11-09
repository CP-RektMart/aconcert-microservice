package handler

import (
	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/dto"
	"github.com/gofiber/fiber/v2"
)

type handler struct {
	domain domain.AuthDomain
}

func NewHandler(domain domain.AuthDomain) *handler {
	return &handler{
		domain: domain,
	}
}

func (h *handler) Mount(r fiber.Router) {
	userGroup := r.Group("/auth")
	userGroup.Post("/login", h.LoginWithProvider)
	userGroup.Post("/refresh", h.RefreshToken)
	userGroup.Post("/logout", h.Logout)
	userGroup.Get("/me", h.GetProfile)
	userGroup.Patch("/me", h.UpdateProfile)
}

func (h *handler) LoginWithProvider(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.LoginWithProviderRequest
	if err := req.Parse(c); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	token, user, isNewUser, err := h.domain.LoginWithGoogle(ctx, req.IdToken)
	if err != nil {
		return errors.Wrap(err, "failed to login with google")
	}

	return c.JSON(dto.LoginWithProviderResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Exp:          token.Exp,
		User:         dto.UserEntityToDTO(user),
		IsNewUser:    isNewUser,
	})
}

func (h *handler) RefreshToken(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.RefreshTokenRequest
	if err := req.Parse(c); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	bearerToken := req.RefreshToken
	token, err := h.domain.RefreshToken(ctx, bearerToken)
	if err != nil {
		return errors.Wrap(err, "failed to refresh token")
	}

	return c.JSON(dto.RefreshTokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Exp:          token.Exp,
	})
}

func (h *handler) Logout(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.LogoutRequest
	if err := req.Parse(c); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := h.domain.Logout(ctx, req.UserID); err != nil {
		return errors.Wrap(err, "failed to logout")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handler) GetProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.GetProfileRequest
	if err := req.Parse(c); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	user, err := h.domain.GetUser(ctx, req.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	return c.JSON(dto.UserEntityToDTO(user))
}

func (h *handler) UpdateProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.UpdateProfileRequest
	if err := req.Parse(c); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	user, err := h.domain.UpdateProfile(ctx, req.UserID, req.Firstname, req.Lastname, req.ProfileImage, req.Birthdate, req.Phone)
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	return c.JSON(dto.UserEntityToDTO(user))
}
