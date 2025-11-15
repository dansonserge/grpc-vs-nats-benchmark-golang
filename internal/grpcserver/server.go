package grpc_server

import (
	"context"

	pb "github.com/dansonserge/grpc-vs-nats-benchmark-golang"
)

// Server implements PingService
type Server struct{}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	// Echo back the payload to keep requests small and deterministic
	return &pb.PingResponse{Payload: req.Payload}, nil
}
