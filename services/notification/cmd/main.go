package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/notification/internal/config"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/handler"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/rabbitmq"
	"github.com/cp-rektmart/aconcert-microservice/pkg/realtime"
)

func main() {
	config := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := logger.Init(config.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}

	rabbitmq.NewRabbitMQConnection(config.RabbitMQ.URL)
	defer rabbitmq.RabbitMQClient.CloseConnection()

	msgs, err := rabbitmq.RabbitMQClient.ConsumeRabbitMQQueue("notifications")
	if err != nil {
		log.Fatalf("Failed to consume RabbitMQ queue: %s", err)
		return
	}

	realtimeService := realtime.New(&config.Realtime)
	domain := domain.New(realtimeService)
	handler := handler.New(domain)

	go func() {
		if err := handler.Mount(ctx, msgs); err != nil {
			log.Fatalf("Failed to mount handler: %s", err)
		}
	}()

	log.Println("🟢 Notification consumer started. Waiting for messages... Press CTRL+C to stop.")
	<-ctx.Done()
	log.Println("🟡 Shutting down consumer gracefully.")
}
