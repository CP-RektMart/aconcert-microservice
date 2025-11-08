package reservation

import (
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/middlewares/authentication"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service        *ReservationService
	authMiddleware authentication.AuthMiddleware
}

func NewHandler(service *ReservationService, authMiddleware authentication.AuthMiddleware) *Handler {
	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}
}

func (h *Handler) Mount(r fiber.Router) {
	group := r.Group("/reservations")
	group.Post("/", h.authMiddleware.Auth, h.CreateReservation)
	group.Delete("/:id", h.authMiddleware.Auth, h.DeleteReservation)
	group.Get("/:id", h.authMiddleware.Auth, h.GetReservation)
	group.Get("/", h.authMiddleware.Auth, h.ListReservation)
	group.Post("/:id/confirm", h.authMiddleware.Auth, h.ConfirmReservation)
}

// @Summary      	Create Reservation
// @Description  	Create a new reservation
// @Tags			reservations
// @Router			/v1/reservations [POST]
// @Security		ApiKeyAuth
// @Param			body	body		dto.CreateReservationRequest	true	"Create reservation request"
// @Success			200 {object}	dto.HttpResponse[dto.CreateReservationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) CreateReservation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.CreateReservationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: "Invalid request body",
		})
	}

	userID, err := h.authMiddleware.GetUserIDFromContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HttpError{
			Error: "UNAUTHORIZED",
		})
	}

	id, err := h.service.CreateReservation(ctx, &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.CreateReservationResponse]{
		Result: dto.CreateReservationResponse{
			ID: id,
		},
	})
}

// @Summary      	Delete Reservation
// @Description  	Delete a reservation
// @Tags			reservations
// @Router			/v1/reservations/{id} [DELETE]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Reservation ID"
// @Success			200 {object}	dto.HttpResponse[dto.DeleteReservationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) DeleteReservation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.DeleteReservationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: "Invalid request parameters",
		})
	}

	id, err := h.service.DeleteReservation(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.DeleteReservationResponse]{
		Result: dto.DeleteReservationResponse{
			ID: id,
		},
	})
}

// @Summary      	Get Reservation
// @Description  	Get a reservation by ID
// @Tags			reservations
// @Router			/v1/reservations/{id} [GET]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Reservation ID"
// @Success			200 {object}	dto.HttpResponse[dto.GetReservationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) GetReservation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.GetReservationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: "Invalid request parameters",
		})
	}

	reservation, err := h.service.GetReservation(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.GetReservationResponse]{
		Result: reservation,
	})
}

// @Summary      	List Reservations
// @Description  	List all reservations for a user
// @Tags			reservations
// @Router			/v1/reservations [GET]
// @Security		ApiKeyAuth
// @Param			userId	query		string	true	"User ID"
// @Success			200 {object}	dto.HttpResponse[dto.ListReservationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) ListReservation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.ListReservationRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: "Invalid query parameters",
		})
	}

	reservations, err := h.service.ListReservation(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.ListReservationResponse]{
		Result: dto.ListReservationResponse{
			Reservations: reservations,
		},
	})
}

// @Summary      	Confirm Reservation
// @Description  	Confirm a reservation
// @Tags			reservations
// @Router			/v1/reservations/{id}/confirm [POST]
// @Security		ApiKeyAuth
// @Param			id		path		string	true	"Reservation ID"
// @Success			200 {object}	dto.HttpResponse[dto.ConfirmReservationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) ConfirmReservation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.ConfirmReservationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: "Invalid request parameters",
		})
	}

	result, err := h.service.ConfirmReservation(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.ConfirmReservationResponse]{
		Result: result,
	})
}
