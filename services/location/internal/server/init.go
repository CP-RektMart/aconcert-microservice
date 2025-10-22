package server

import (
	"context"
	"time"

	"github.com/cp-rektmart/aconcert-microservice/location/internal/repository"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	locationpb "github.com/cp-rektmart/aconcert-microservice/pkg/proto/location"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type LocationService struct {
	locationpb.UnimplementedLocationServiceServer
	locationRepo *repository.LocationRepository
}

func NewLocationService(db *mongo.Database) *LocationService {
	locationRepo := repository.NewLocationRepository(db, "locations")

	return &LocationService{
		locationRepo: locationRepo,
	}
}

func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		logger.InfoContext(ctx, "gRPC request received",
			"method", info.FullMethod,
			"request", req,
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		if err != nil {
			logger.ErrorContext(ctx, "gRPC request failed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", st.Message(),
				"code", st.Code().String(),
			)
		} else {
			logger.InfoContext(ctx, "gRPC request completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return resp, err
	}
}
