package main

import (
	"log"
	"net"
	"os"
	"strconv"

	_ "google.golang.org/genproto/protobuf/ptype"

	grpcserver "github.com/cp-rektmart/aconcert-microservice/event/internal/grpc-server"
	pb "github.com/cp-rektmart/aconcert-microservice/event/internal/proto"
	"github.com/cp-rektmart/aconcert-microservice/event/internal/store"
	"github.com/cp-rektmart/aconcert-microservice/pkg/logger"
	"google.golang.org/grpc"
)

func main() {
	addr := ":" + strconv.Itoa(50051)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Panic("failed to listen: %v", err)
	}

	store, err := store.NewStore(os.Getenv("DATABASE_URL"), os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}

	grpcServer := grpc.NewServer()
	eventServer := grpcserver.NewEventServer(store)

	pb.RegisterEventServiceServer(grpcServer, eventServer)

	log.Println("gRPC server running on", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
