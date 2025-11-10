package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/pkg/grpclogger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/postgres"
	reservationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/reservation"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"github.com/cp-rektmart/aconcert-microservice/reservation/config"
	db "github.com/cp-rektmart/aconcert-microservice/reservation/db/codegen"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/domains"
	"github.com/cp-rektmart/aconcert-microservice/reservation/internal/repositories"
	"google.golang.org/grpc"
)

// grpc server
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

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	if err != nil {
		logger.PanicContext(ctx, "failed to listen: %v", err)
	}

	queries := db.New(pgConn)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpclogger.LoggingUnaryInterceptor),
	)

	reservationRepo := repositories.NewReservationRepository(queries, pgConn, redisConn)

	reservationServer := domains.New(reservationRepo, conf.Stripe)
	reservationpb.RegisterReservationServiceServer(grpcServer, reservationServer)

	// Start Redis expiration listener in background
	go func() {
		reservationRepo.StartExpirationListener(ctx)
	}()

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

	// Gracefully stop gRPC
	grpcServer.GracefulStop()
	logger.InfoContext(ctx, "gRPC server stopped cleanly")
}
