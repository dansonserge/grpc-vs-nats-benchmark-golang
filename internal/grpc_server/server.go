package grpc_server

import (
	"context"
	pb "grpc-vs-nats-benchmark-golang/proto"
)

type Server struct {
	pb.UnimplementedPingServiceServer
}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Payload: req.Payload}, nil
}
