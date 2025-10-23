package event

import (
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/middlewares/authentication"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service        *EventService
	authentication authentication.AuthMiddleware
}

func NewHandler(service *EventService, authentication authentication.AuthMiddleware) *Handler {
	return &Handler{
		service:        service,
		authentication: authentication,
	}
}

func (h *Handler) Mount(r fiber.Router) {
	group := r.Group("/events")
	group.Get("/", h.ListEvents)
	group.Get("/:id", h.GetEvent)
	group.Post("/", h.authentication.Auth, h.CreateEvent)
	group.Put("/", h.authentication.Auth, h.UpdateEvent)
	group.Delete("/", h.authentication.Auth, h.DeleteEvent)
}

// @Summary      	List Events
// @Description  	List Events
// @Tags			events
// @Router			/v2/events [GET]
// @Param			query		query		int			false	"query"
// @Param			sortBy		query		int			false	"sortBy"
// @Param			order		query		string		false	"order"
// @Param			page		query		string		false	"page"
// @Param			limit		query		string		false	"limit"
// @Success			200 {object}	dto.HttpResponse[dto.EventListResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) ListEvents(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.ListEvents
	if err := c.QueryParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	events, err := h.service.ListEvents(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.EventListResponse]{
		Result: dto.EventListResponse{
			List: events,
		},
	})
}

// @Summary      	Get Event
// @Description  	Get Event
// @Tags			events
// @Router			/v2/events/{id} [GET]
// @Param			id	path		string	true	"Event ID"
// @Success			200 {object}	dto.HttpResponse[dto.EventResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) GetEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.GetEvent
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	event, err := h.service.GetEvent(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.EventResponse]{
		Result: event,
	})
}

// @Summary      	Create Event
// @Description  	Create Event
// @Tags			events
// @Router			/v2/events [POST]
// @Param			body	body		dto.CreateEvent	true	"Create event request"
// @Success			200 {object}	dto.HttpResponse[string]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) CreateEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.CreateEvent
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	id, err := h.service.CreateEvent(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[string]{
		Result: id,
	})
}

// @Summary      	Update Event
// @Description  	Update Event
// @Tags			events
// @Router			/v2/events/{id} [PUT]
// @Param			id	path		string	true	"Event ID"
// @Param			body	body		dto.UpdateEvent	true	"Update event request"
// @Success			200 {object}	dto.HttpResponse[string]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) UpdateEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.UpdateEvent
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	id, err := h.service.UpdateEvent(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[string]{
		Result: id,
	})
}

// @Summary      	Delete Event
// @Description  	Delete Event
// @Tags			events
// @Router			/v2/events/{id} [DELETE]
// @Param			id	path		string	true	"Event ID"
// @Success			204
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) DeleteEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.DeleteEvent
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := h.service.DeleteEvent(ctx, &req); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
