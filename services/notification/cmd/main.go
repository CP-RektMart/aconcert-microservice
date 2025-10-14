package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/notification/internal/config"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/handler"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/hub"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/repository"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/rabbitmq"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// LOAD ENVIRONMENT VARIABLES
	config := config.Load()
	// INITIALIZE LOGGER
	if err := logger.Init(config.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}

	// CONNECT TO RABBITMQ
	rabbitmq.NewRabbitMQConnection(config.RabbitMQ.URL)
	// DEFER CLOSE CONNECTION TO RABBITMQ
	defer rabbitmq.RabbitMQClient.CloseConnection()

	// CONSUME RABBITMQ QUEUE --> "notifications"
	msgs, err := rabbitmq.RabbitMQClient.ConsumeRabbitMQQueue("notifications")

	if err != nil {
		log.Fatalf("Failed to consume RabbitMQ queue: %s", err)
		return
	}

	redisConn, err := redis.New(ctx, config.Redis)
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to redis", slog.Any("error", err))
	}
	defer func() {
		if err := redisConn.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close redis connection", slog.Any("error", err))
		}
	}()

	app := fiber.New(fiber.Config{
		AppName:       config.Name,
		BodyLimit:     config.MaxBodyLimit * 1024 * 1024,
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
		AllowOrigins:     config.Cors.AllowedOrigins,
		AllowMethods:     config.Cors.AllowedMethods,
		AllowHeaders:     config.Cors.AllowedHeaders,
		AllowCredentials: config.Cors.AllowCredentials,
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

	// CHANNEL TO RECEIVE NOTIFICATIONS
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var eventData entities.EventData
			err := json.Unmarshal(d.Body, &eventData)
			if err != nil {
				log.Printf("Error reading event data (please check the JSON format): %s", err)
				continue
			}

			domain.PushMessage(ctx, eventData.UserID, eventData.EventType, eventData.Data)
		}
	}()

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", config.Port)); err != nil {
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

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
