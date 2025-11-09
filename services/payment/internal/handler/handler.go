package handler

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/payment/internal/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v83/webhook"
)

type Handler struct {
	domain              *domain.Domain
	stripeSigningSecret string
}

func NewHandler(domain *domain.Domain, stripeSigningSecret string) *Handler {
	return &Handler{
		domain:              domain,
		stripeSigningSecret: stripeSigningSecret,
	}
}

func (h *Handler) Mount(r fiber.Router) {
	r.Post("/stripe/webhook", h.HandleStripeWebhook)
}

func (h *Handler) HandleStripeWebhook(c *fiber.Ctx) error {
	ctx := c.UserContext()

	const MaxBodyBytes = int64(65536)

	// Limit body size
	if len(c.Body()) > int(MaxBodyBytes) {
		return c.SendStatus(fiber.StatusRequestEntityTooLarge)
	}

	payload := c.Body()
	sigHeader := c.Get("Stripe-Signature")

	event, err := webhook.ConstructEvent(payload, sigHeader, h.stripeSigningSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Signature verification failed: %v\n", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := h.domain.ProcessPayment(ctx, event); err != nil {
		return errors.Wrap(err, "can't process this event")
	}

	return c.SendStatus(204)
}
