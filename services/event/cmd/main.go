package main

import (
	"log"
	"net"

	_ "google.golang.org/genproto/protobuf/ptype"

	grpcserver "github.com/cp-rektmart/aconcert-microservice/event/grpc-server"
	pb "github.com/cp-rektmart/aconcert-microservice/event/proto"
	"google.golang.org/grpc"
)

func main() {
	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	eventServer := grpcserver.NewEventServer()
	pb.RegisterEventServiceServer(grpcServer, eventServer)

	log.Println("gRPC server running on", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
