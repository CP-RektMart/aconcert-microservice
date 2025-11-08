package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/payment/internal/config"
	"github.com/cp-rektmart/aconcert-microservice/payment/internal/domain"
	"github.com/cp-rektmart/aconcert-microservice/payment/internal/handler"
	"github.com/cp-rektmart/aconcert-microservice/payment/internal/repository"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/pkg/requestlogger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/stripe/stripe-go/v79"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conf := config.Load()
	stripe.Key = conf.Stripe.SecretKey

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := logger.Init(conf.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}

	reservationConn, err := grpc.NewClient(conf.ReservationClientBaseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to reservation service", slog.Any("error", err))
	}
	reservationClient := reservationpb.NewReservationServiceClient(reservationConn)

	repo := repository.NewRepository(reservationClient)
	domain := domain.NewService(repo)
	handler := handler.NewHandler(domain, conf.Stripe.SigningSecret)

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

	app.Use(requestid.New()).Use(requestlogger.New())

	v1 := app.Group("/v1")
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
