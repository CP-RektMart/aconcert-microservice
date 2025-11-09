package handler

import (
	"bufio"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/pkg/requestlogger"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/hub"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/valyala/fasthttp"
)

type handler struct {
	hub    *hub.Hub
	domain *domain.Domain
}

func New(h *hub.Hub, domain *domain.Domain) *handler {
	return &handler{
		hub:    h,
		domain: domain,
	}
}

func (h *handler) Mount(r fiber.Router) {
	sseGroup := r.Group("/", requestid.New(), requestlogger.New())
	sseGroup.Get("/realtime", h.Realtime) // From Frontend

	httpGroup := r.Group("/", otelfiber.Middleware(), requestid.New(), requestlogger.New())
	httpGroup.Post("/push-message", h.PushMessage) // From Internal Only
	httpGroup.Post("/ack", h.Ack)                  // From Frontend
}

func (h *handler) Realtime(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req dto.RealtimeRequest
	if err := req.Parse(c); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	// 2) SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no") // disable nginx buffering
	c.Status(fiber.StatusOK)

	// 3) open stream writer
	client := make(hub.Client, 10)
	h.hub.Register(ctx, req.UserID, client)

	if req.State == "reconnect" {
		h.domain.ResendCache(ctx, req.UserID)
	}

	clientGone := c.Context().Done()

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// ensure we unregister at exit
		defer h.hub.Unregister(ctx, req.UserID, client)

		// **initial ping** to flush headers
		fmt.Fprint(w, ": connected\n\n")
		if err := w.Flush(); err != nil {
			return
		}

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		// 4) main loop
		for {
			select {
			case msg, ok := <-client:
				if !ok {
					return
				}
				// SSE event + data
				fmt.Fprintf(w, "event: %s\n", msg.EventType)
				fmt.Fprintf(w, "data: %s\n\n", msg.Data)
				if err := w.Flush(); err != nil {
					return
				}

			case <-ticker.C:
				fmt.Fprintf(w, "event: PING\n")
				fmt.Fprintf(w, "data: ping\n\n")
				w.Flush()

			case <-clientGone:
				return
			}
		}
	}))

	return nil
}

func (h *handler) PushMessage(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.PushMessage
	if err := req.Parse(c); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	if err := h.domain.PushMessage(ctx, req.UserID, req.EventType, req.Data); err != nil {
		return errors.Wrap(err, "failed to push message")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handler) Ack(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.AckPushMessage
	if err := req.Parse(c); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HttpError{
			Error: err.Error(),
		})
	}

	if err := h.domain.ReceiveAck(ctx, req.EventID); err != nil {
		return errors.Wrap(err, "failed to ack push message")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
