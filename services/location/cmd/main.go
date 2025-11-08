package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/cp-rektmart/aconcert-microservice/location/config"
	"github.com/cp-rektmart/aconcert-microservice/location/internal/server"
	"github.com/cp-rektmart/aconcert-microservice/pkg/grpclogger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/mongodb"
	locationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/location"
	"google.golang.org/grpc"
)

func main() {
	conf := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := logger.Init(conf.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}

	mongoClient, err := mongodb.NewMongo(mongodb.Config{
		Host:     conf.Mongo.Host,
		Port:     conf.Mongo.Port,
		User:     conf.Mongo.User,
		Password: conf.Mongo.Password,
		Database: conf.Mongo.Database,
	})
	if err != nil {
		logger.Panic("failed to connect to MongoDB", slog.Any("error", err))
	}
	defer mongoClient.Disconnect(ctx)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	if err != nil {
		logger.PanicContext(ctx, "failed to listen: %v", err)
	}

	// gRPC server setup
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpclogger.LoggingUnaryInterceptor),
	)

	locationService := server.NewLocationService(mongoClient.DB)
	locationpb.RegisterLocationServiceServer(grpcServer, locationService)

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
