package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "google.golang.org/genproto/protobuf/ptype"

	db "github.com/cp-rektmart/aconcert-microservice/event/db/codegen"
	"github.com/cp-rektmart/aconcert-microservice/event/internal/config"
	eventService "github.com/cp-rektmart/aconcert-microservice/event/internal/service"
	"github.com/cp-rektmart/aconcert-microservice/pkg/grpclogger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/postgres"
	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
	"github.com/cp-rektmart/aconcert-microservice/pkg/rabbitmq"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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

	redisClient, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logger.PanicContext(ctx, "failed to connect to redis", slog.Any("error", err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close redis connection", slog.Any("error", err))
		}
	}()

	rabbitmq.NewRabbitMQConnection(conf.RabbitMQ.URL)
	defer rabbitmq.RabbitMQClient.CloseConnection()

	queries := db.New(pgConn)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	if err != nil {
		logger.PanicContext(ctx, "failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpclogger.LoggingUnaryInterceptor),
	)

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Set service as serving
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("event.EventService", healthpb.HealthCheckResponse_SERVING)

	eventServ := eventService.NewEventService(queries)
	eventpb.RegisterEventServiceServer(grpcServer, eventServ)

	go func() {
		logger.InfoContext(ctx, "starting gRPC server", slog.String("port", strconv.Itoa(conf.Port)))
		if err := grpcServer.Serve(lis); err != nil {
			logger.PanicContext(ctx, "failed to serve", slog.Any("error", err))
			stop() // stop context if server fails
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	logger.InfoContext(ctx, "shutting down gRPC server gracefully")

	// Mark as not serving before shutdown
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	// Gracefully stop gRPC
	grpcServer.GracefulStop()
	logger.InfoContext(ctx, "gRPC server stopped cleanly")
}
