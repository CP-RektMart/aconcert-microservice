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
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/notification/internal/handler"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/rabbitmq"
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

	domain := domain.New()
	handler := handler.New(domain)

	go func() {
		for d := range msgs {
			var eventData entities.Message
			if err := json.Unmarshal(d.Body, &eventData); err != nil {
				log.Printf("Error reading event data (invalid JSON): %s", err)
				continue
			}

			fmt.Printf("📩 Received message: %+v\n", eventData)
			switch eventData.Type {
			case entities.MessageTypeEventCreated:
				if err := handler.HandleEventCreated(ctx); err != nil {
					log.Printf("Error handling event.created: %s", err)
				}
			case entities.MessageTypeEventUpdated:
				if err := handler.HandleEventUpdated(ctx); err != nil {
					log.Printf("Error handling event.updated: %s", err)
				}
			case entities.MessageTypeReservationConfirmed:
				if err := handler.HandleReservationConfirmed(ctx); err != nil {
					log.Printf("Error handling reservation.confirmed: %s", err)
				}
			case entities.MessageTypeReservationCancelled:
				if err := handler.HandleReservationCancelled(ctx); err != nil {
					log.Printf("Error handling reservation.cancelled: %s", err)
				}
			default:
				log.Printf("Unknown message type: %s", eventData.Type)
				continue
			}
		}
	}()

	log.Println("🟢 Notification consumer started. Waiting for messages... Press CTRL+C to stop.")
	<-ctx.Done()
	log.Println("🟡 Shutting down consumer gracefully.")
}
