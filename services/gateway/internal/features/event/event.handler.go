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
	group.Put("/:id", h.authentication.Auth, h.UpdateEvent)
	group.Delete("/:id", h.authentication.Auth, h.DeleteEvent)

	group.Get("/:id/event-zones", h.GetEventZoneByEventID)
	group.Post("/:id/event-zones", h.authentication.Auth, h.CreateEventZone)
	group.Put("/:id/event-zones/:zoneId", h.authentication.Auth, h.UpdateEventZone)
	group.Delete("/:id/event-zones/:zoneId", h.authentication.Auth, h.DeleteEventZone)
}

// @Summary      	List Events
// @Description  	List Events
// @Tags			events
// @Router			/v1/events [GET]
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

	var req dto.ListEventsRequest
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
// @Router			/v1/events/{id} [GET]
// @Param			id	path		string	true	"Event ID"
// @Success			200 {object}	dto.HttpResponse[dto.EventResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) GetEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.GetEventRequest
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
// @Router			/v1/events [POST]
// @Security		ApiKeyAuth
// @Param			body	body		dto.CreateEventRequest	true	"Create event request"
// @Success			200 {object}	dto.HttpResponse[dto.CreateEventResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) CreateEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	id, err := h.service.CreateEvent(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.CreateEventResponse]{
		Result: dto.CreateEventResponse{
			ID: id,
		},
	})
}

// @Summary      	Update Event
// @Description  	Update Event
// @Tags			events
// @Router			/v1/events/{id} [PUT]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Event ID"
// @Param			body	body		dto.UpdateEventRequest	true	"Update event request"
// @Success			200 {object}	dto.HttpResponse[dto.UpdateEventResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) UpdateEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.UpdateEventRequest
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

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.UpdateEventResponse]{
		Result: dto.UpdateEventResponse{
			ID: id,
		},
	})
}

// @Summary      	Delete Event
// @Description  	Delete Event
// @Tags			events
// @Router			/v1/events/{id} [DELETE]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Event ID"
// @Success			204
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) DeleteEvent(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.DeleteEventRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := h.service.DeleteEvent(ctx, &req); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary      	Get Event Zones by Event ID
// @Description  	Get Event Zones by Event ID
// @Tags			event-zones
// @Router			/v1/event-zones/event/{eventId} [GET]
// @Param			eventId	path		string	true	"Event ID"
// @Success			200 {object}	dto.HttpResponse[dto.EventZoneListResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) GetEventZoneByEventID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.GetEventZoneByEventIDRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	zones, err := h.service.GetEventZonesByEventID(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.EventZoneListResponse]{
		Result: dto.EventZoneListResponse{
			List: zones,
		},
	})
}

// @Summary      	Create Event Zone
// @Description  	Create Event Zone
// @Tags			event-zones
// @Router			/v1/event-zones [POST]
// @Security		ApiKeyAuth
// @Param			body	body		dto.CreateEventZoneRequest	true	"Create event zone request"
// @Success			200 {object}	dto.HttpResponse[dto.CreateEventZoneResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) CreateEventZone(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.CreateEventZoneRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	id, err := h.service.CreateEventZone(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.CreateEventZoneResponse]{
		Result: dto.CreateEventZoneResponse{
			ID: id,
		},
	})
}

// @Summary      	Update Event Zone
// @Description  	Update Event Zone
// @Tags			event-zones
// @Router			/v1/event-zones/{id} [PUT]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Event Zone ID"
// @Param			body	body		dto.UpdateEventZoneRequest	true	"Update event zone request"
// @Success			200 {object}	dto.HttpResponse[dto.UpdateEventZoneResponse]
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) UpdateEventZone(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.UpdateEventZoneRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	id, err := h.service.UpdateEventZone(ctx, &req)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(dto.HttpResponse[dto.UpdateEventZoneResponse]{
		Result: dto.UpdateEventZoneResponse{
			ID: id,
		},
	})
}

// @Summary      	Delete Event Zone
// @Description  	Delete Event Zone
// @Tags			event-zones
// @Router			/v1/event-zones/{id} [DELETE]
// @Security		ApiKeyAuth
// @Param			id	path		string	true	"Event Zone ID"
// @Success			204
// @Failure			400	{object}	dto.HttpError
// @Failure			500	{object}	dto.HttpError
func (h *Handler) DeleteEventZone(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.DeleteEventZoneRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := h.service.DeleteEventZone(ctx, &req); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
