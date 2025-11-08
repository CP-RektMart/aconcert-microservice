package location

import (
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/middlewares/authentication"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service        *LocationService
	authentication authentication.AuthMiddleware
}

func NewHandler(service *LocationService, authentication authentication.AuthMiddleware) *Handler {
	return &Handler{
		service:        service,
		authentication: authentication,
	}
}

func (h *Handler) Mount(r fiber.Router) {
	group := r.Group("/locations")
	group.Get("/", h.ListLocations)
	group.Get("/:id", h.GetLocation)
	group.Post("/", h.authentication.Auth, h.CreateLocation)
	group.Put("/:id", h.authentication.Auth, h.UpdateLocation)
	group.Delete("/:id", h.authentication.Auth, h.DeleteLocation)
}

// @Summary      	List Locations
// @Description  	List Locations
// @Tags			locations
// @Router			/v1/locations [GET]
// @Success			200 {object}	dto.HttpResponse[dto.ListLocationsResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) ListLocations(c *fiber.Ctx) error {
	ctx := c.UserContext()

	locations, err := h.service.ListLocations(ctx, &dto.ListLocationsRequest{})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.ListLocationsResponse]{
		Result: locations,
	})
}

// @Summary      	Get Location
// @Description  	Get Location
// @Tags			locations
// @Router			/v1/locations/{id} [GET]
// @Param			id	path		string	true	"Location ID"
// @Success			200 {object}	dto.HttpResponse[dto.LocationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) GetLocation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.GetLocationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	location, err := h.service.GetLocation(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.LocationResponse]{
		Result: location,
	})
}

// @Summary      	Create Location
// @Description  	Create Location
// @Tags			locations
// @Router			/v1/locations [POST]
// @Security		ApiKeyAuth
// @Param			body	body		dto.CreateLocationRequest	true	"Create location request"
// @Success			200 {object}	dto.HttpResponse[dto.CreateLocationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) CreateLocation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.CreateLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	response, err := h.service.CreateLocation(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.CreateLocationResponse]{
		Result: response,
	})
}

// @Summary      	Update Location
// @Description  	Update Location
// @Tags			locations
// @Router			/v1/locations/{id} [PUT]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Location ID"
// @Param			body	body		dto.UpdateLocationRequest	true	"Update location request"
// @Success			200 {object}	dto.HttpResponse[dto.UpdateLocationResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) UpdateLocation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.UpdateLocationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	response, err := h.service.UpdateLocation(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.UpdateLocationResponse]{
		Result: response,
	})
}

// @Summary      	Delete Location
// @Description  	Delete Location
// @Tags			locations
// @Router			/v1/locations/{id} [DELETE]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Location ID"
// @Success			204
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) DeleteLocation(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.DeleteLocationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := h.service.DeleteLocation(ctx, &req); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
