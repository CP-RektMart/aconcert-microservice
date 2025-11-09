package auth

import (
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/middlewares/authentication"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service        *AuthService
	authentication authentication.AuthMiddleware
}

func NewHandler(service *AuthService, authentication authentication.AuthMiddleware) *Handler {
	return &Handler{
		service:        service,
		authentication: authentication,
	}
}

func (h *Handler) Mount(r fiber.Router) {
	group := r.Group("/auth")
	group.Post("/login", h.Login)
	group.Post("/refresh", h.RefreshToken)
	group.Post("/logout", h.authentication.Auth, h.Logout)
	group.Get("/me", h.authentication.Auth, h.GetProfile)
	group.Patch("/me", h.authentication.Auth, h.UpdateProfile)
}

// @Summary			Login
// @Description		Login
// @Tags			auth
// @Router			/v1/auth/login [POST]
// @Param 			RequestBody 	body 	dto.LoginRequest 	true 	"request request"
// @Success			200	{object}	dto.HttpResponse[dto.LoginResponse]
// @Failure			500	{object}	dto.HttpError
func (h *Handler) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	response, err := h.service.Login(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.LoginResponse]{
		Result: response,
	})
}

// @Summary			Refresh Token
// @Description		Refresh Token
// @Tags			auth
// @Router			/v1/auth/refresh [POST]
// @Param 			RequestBody 	body 	dto.RefreshTokenRequest 	true 	"request request"
// @Success			200	{object}	dto.HttpResponse[dto.RefreshTokenResponse]
// @Failure			401	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	response, err := h.service.RefreshToken(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.RefreshTokenResponse]{
		Result: response,
	})
}

// @Summary			Logout
// @Description		Logout
// @Tags			auth
// @Router			/v1/auth/logout [POST]
// @Security		ApiKeyAuth
// @Success			204
// @Failure			401	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) Logout(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, err := h.authentication.GetUserIDFromContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HttpError{
			Error: "UNAUTHORIZED",
		})
	}

	if err := h.service.Logout(ctx, &dto.LogoutRequest{
		UserID: userID,
	}); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary			Get Profile
// @Description		Get Profile
// @Tags			auth
// @Router			/v1/auth/me [GET]
// @Security		ApiKeyAuth
// @Success			200	{object}	dto.HttpResponse[dto.UserResponse]
// @Failure			401	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, err := h.authentication.GetUserIDFromContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HttpError{
			Error: "UNAUTHORIZED",
		})
	}

	response, err := h.service.GetProfile(ctx, &dto.GetProfileRequest{
		UserID: userID,
	})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.UserResponse]{
		Result: response,
	})
}

// @Summary			Update Profile
// @Description		Update Profile
// @Tags			auth
// @Router			/v1/auth/me [PATCH]
// @Security		ApiKeyAuth
// @Param 			RequestBody 	body 	dto.UpdateProfileRequest 	true 	"request request"
// @Success			200	{object}	dto.HttpResponse[dto.UserResponse]
// @Failure			401	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, err := h.authentication.GetUserIDFromContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HttpError{
			Error: "UNAUTHORIZED",
		})
	}

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if _, err := h.service.UpdateProfile(ctx, &dto.UpdateProfileRequest{
		UserID:       userID,
		Firstname:    req.Firstname,
		Lastname:     req.Lastname,
		ProfileImage: req.ProfileImage,
		Birthdate:    req.Birthdate,
		Phone:        req.Phone,
	}); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
