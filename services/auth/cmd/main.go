package main

import "fmt"

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	
	// Register ProductService server
	pb.RegisterProductServiceServer(grpcServer, &ProductServer{})

	log.Println("gRPC server running on", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}	
}
