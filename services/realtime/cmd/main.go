package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/config"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/handler"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/hub"
	"github.com/cp-rektmart/aconcert-microservice/realtime/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	conf := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := logger.Init(conf.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}

	// Init Redis
	redisConn, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to redis", slog.Any("error", err))
	}
	defer func() {
		if err := redisConn.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close redis connection", slog.Any("error", err))
		}
	}()

	app := fiber.New(fiber.Config{
		AppName:       conf.Name,
		BodyLimit:     conf.MaxBodyLimit * 1024 * 1024,
		CaseSensitive: true,
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.ErrorContext(c.UserContext(), "unhandled error", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(dto.HttpError{
				Error: "Internal Server Error",
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     conf.Cors.AllowedOrigins,
		AllowMethods:     conf.Cors.AllowedMethods,
		AllowHeaders:     conf.Cors.AllowedHeaders,
		AllowCredentials: conf.Cors.AllowCredentials,
	}))

	hub := hub.New()
	repo := repository.New(redisConn)
	domain := domain.New(hub, repo)
	handler := handler.New(hub, domain)

	v1 := app.Group("/v1")
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(dto.HttpResponse[string]{
			Result: "ok",
		})
	})

	handler.Mount(v1)

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", conf.Port)); err != nil {
			logger.PanicContext(ctx, "failed to start server", slog.Any("error", err))
			stop()
		}
	}()

	defer func() {
		if err := app.ShutdownWithContext(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown server", slog.Any("error", err))
		}
		logger.InfoContext(ctx, "gracefully shutdown server")
	}()

	<-ctx.Done()
	logger.InfoContext(ctx, "Shutting down server")
}
