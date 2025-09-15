package requestlogger

import (
	"log/slog"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func New() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		requestID := ctx.GetRespHeader("X-Request-Id")
		err := ctx.Next()
		if err != nil {
			return errors.Wrap(err, "can't get request id")
		}

		logger.InfoContext(ctx.UserContext(), "request received", slog.String("request_id", requestID), slog.String("method", ctx.Method()), slog.String("path", ctx.Path()), slog.Int("status", ctx.Response().StatusCode()))
		return nil
	}
}
