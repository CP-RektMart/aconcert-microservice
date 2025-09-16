package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/auth/config"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/postgres"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/cp-rektmart/aconcert-microservice/pkg/requestlogger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	conf := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := logger.Init(conf.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}

	pgConn, err := postgres.NewPool(ctx, conf.Postgres)
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to postgres", slog.Any("error", err))
	}
	defer pgConn.Close()

	redisConn, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to redis", slog.Any("error", err))
	}
	defer func() {
		if err := redisConn.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close redis connection", slog.Any("error", err))
		}
	}()

	// queries := gen.New(pgConn)

	app := fiber.New(fiber.Config{
		AppName:       conf.Name,
		BodyLimit:     conf.MaxBodyLimit * 1024 * 1024,
		CaseSensitive: true,
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.ErrorContext(c.UserContext(), "unhandled error", slog.Any("error", err))
			return c.SendStatus(fiber.StatusInternalServerError)
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     conf.Cors.AllowedOrigins,
		AllowMethods:     conf.Cors.AllowedMethods,
		AllowHeaders:     conf.Cors.AllowedHeaders,
		AllowCredentials: conf.Cors.AllowCredentials,
	})).
		Use(requestid.New()).
		Use(requestlogger.New())

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
