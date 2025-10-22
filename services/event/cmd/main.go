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

	"github.com/cp-rektmart/aconcert-microservice/event/config"
	db "github.com/cp-rektmart/aconcert-microservice/event/db/codegen"
	eventService "github.com/cp-rektmart/aconcert-microservice/event/internal/service/event"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/postgres"
	eventpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/event"
	"github.com/cp-rektmart/aconcert-microservice/pkg/redis"
	"google.golang.org/grpc"
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

	queries := db.New(pgConn)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	if err != nil {
		logger.PanicContext(ctx, "failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	eventServ := eventService.NewEventService(queries)
	eventpb.RegisterEventServiceServer(grpcServer, eventServ)

	logger.InfoContext(ctx, "starting gRPC server", slog.String("port", strconv.Itoa(conf.Port)))

	if err := grpcServer.Serve(lis); err != nil {
		logger.PanicContext(ctx, "failed to serve: %v", slog.Any("error", err))
	}
}
