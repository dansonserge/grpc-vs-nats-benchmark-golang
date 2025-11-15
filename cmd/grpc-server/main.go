package main

import (
	"flag"
	"log"
	"net"

	"grpc-vs-nats-benchmark-golang/internal/grpc_server"

	pb "grpc-vs-nats-benchmark-golang/proto"

	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":50051", "gRPC listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPingServiceServer(s, &grpc_server.Server{})

	log.Printf("gRPC server listening on %s", *addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve error: %v", err)
	}
}
