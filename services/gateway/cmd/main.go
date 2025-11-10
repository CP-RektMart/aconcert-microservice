package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/gateway/docs"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/config"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/dto"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/features/auth"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/features/event"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/features/location"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/features/reservation"
	"github.com/cp-rektmart/aconcert-microservice/gateway/internal/middlewares/authentication"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
	locationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/location"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/cp-rektmart/aconcert-microservice/pkg/requestlogger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"github.com/swaggo/swag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title						A Concert Gateway
// @version						1.0.0
// @description					A Concert Gateway API Documentation
// @securityDefinitions.apikey ApiKeyAuth
// @in							header
// @name						Authorization
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	conf := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := logger.Init(conf.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}

	authRedisClient, err := redis.New(ctx, conf.AuthRedis)
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to redis", slog.Any("error", err))
	}
	defer func() {
		if err := authRedisClient.Close(); err != nil {
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
	app.Use(healthcheck.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     conf.Cors.AllowedOrigins,
		AllowMethods:     conf.Cors.AllowedMethods,
		AllowHeaders:     conf.Cors.AllowedHeaders,
		AllowCredentials: conf.Cors.AllowCredentials,
	})).
		Use(requestid.New()).
		Use(requestlogger.New())

	authMiddleware := authentication.NewAuthMiddleware(&conf.JWT, authRedisClient)

	authService := auth.NewService(conf.AuthClientBaseURL)
	authHandler := auth.NewHandler(authService, authMiddleware)

	eventConn, err := grpc.NewClient(conf.EventClientBaseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to event service", slog.Any("error", err))
	}
	eventClient := eventpb.NewEventServiceClient(eventConn)
	eventService := event.NewService(eventClient)
	eventHandler := event.NewHandler(eventService, authMiddleware)

	locationConn, err := grpc.NewClient(conf.LocationClientBaseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to location service", slog.Any("error", err))
	}
	locationClient := locationpb.NewLocationServiceClient(locationConn)
	locationService := location.NewService(locationClient)
	locationHandler := location.NewHandler(locationService, authMiddleware)

	reservationConn, err := grpc.NewClient(conf.ReservationClientBaseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to reservation service", slog.Any("error", err))
	}
	reservationClient := reservationpb.NewReservationServiceClient(reservationConn)
	reservationService := reservation.NewService(reservationClient)
	reservationHandler := reservation.NewHandler(reservationService, authMiddleware)

	v1 := app.Group("/v1")
	authHandler.Mount(v1)
	eventHandler.Mount(v1)
	locationHandler.Mount(v1)
	reservationHandler.Mount(v1)

	swag.Register(docs.SwaggerInfo.InfoInstanceName, docs.SwaggerInfo)
	if conf.Environment != "production" {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

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
