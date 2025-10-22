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
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"github.com/cp-rektmart/aconcert-microservice/pkg/mongodb"
	locationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/location"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	conf := config.Load()

	// Context with signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize logger
	if err := logger.Init(conf.Logger); err != nil {
		logger.PanicContext(ctx, "failed to initialize logger", slog.Any("error", err))
	}
	logger.Info("Logger initialized")

	// service startup
	logger.Info(conf.Name+" starting...", slog.String("environment", conf.Environment))
	defer logger.Info(conf.Name + " stopped")

	// Connect to MongoDB
	logger.Info("Connecting to MongoDB...",
		slog.String("host", conf.Mongo.Host),
		slog.Int("port", conf.Mongo.Port),
		slog.String("database", conf.Mongo.Database),
	)
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
	logger.Info("Connected to MongoDB")
	defer func() {
		logger.Info("Disconnecting MongoDB...")
		mongoClient.Disconnect(ctx)
		logger.Info("MongoDB disconnected")
	}()

	// gRPC server setup
	logger.Info("Starting gRPC server...", slog.Int("port", conf.Port))
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(server.UnaryLoggingInterceptor()),
	)
	locationService := server.NewLocationService(mongoClient.DB)
	locationpb.RegisterLocationServiceServer(grpcServer, locationService)

	listenAddr := ":" + strconv.Itoa(conf.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Panic("failed to listen", slog.Any("error", err))
	}
	logger.Info("gRPC server listening", slog.String("port", strconv.Itoa(conf.Port)))

	defer func() {
		logger.Info("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		logger.Info("gRPC server stopped")
	}()

	// Handle graceful stop on interrupt signal
	go func() {
		<-ctx.Done()
		logger.Info("Shutdown signal received")
		grpcServer.GracefulStop()
	}()

	// Start serving
	if err := grpcServer.Serve(lis); err != nil {
		logger.Panic("failed to serve gRPC", slog.Any("error", err))
	}
}
