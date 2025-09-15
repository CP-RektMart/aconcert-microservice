package handler

import (
	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/middlewares/authentication"
	"github.com/gofiber/fiber/v2"
)

type handler struct {
	domain         domain.Domain
	authMiddleware authentication.AuthMiddleware
}

func NewHandler(domain domain.Domain, authMiddleware authentication.AuthMiddleware) *handler {
	return &handler{
		domain:         domain,
		authMiddleware: authMiddleware,
	}
}

func (h *handler) Mount(r fiber.Router) {
	userGroup := r.Group("/auth")
	userGroup.Post("/login", h.LoginWithProvider)
	// userGroup.Post("/refresh", h.RefreshToken)
	// userGroup.Post("/logout", h.authMiddleware.Auth, h.Logout)
	// userGroup.Get("/me", h.authMiddleware.Auth, h.GetProfile)
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
