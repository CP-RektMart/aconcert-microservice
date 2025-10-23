package grpclogger

import (
	"context"
	"log/slog"
	"time"

	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	res, err := handler(ctx, req)
	elapsed := time.Since(start)

	st, _ := status.FromError(err)
	logger.InfoContext(ctx,
		"gRPC request",
		slog.String("method", info.FullMethod),
		slog.Duration("latency", elapsed),
		slog.String("status", st.Code().String()),
	)

	return res, err
}
