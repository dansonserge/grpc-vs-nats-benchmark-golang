package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"grpc-vs-nats-benchmark-golang/internal/bench"
	pb "grpc-vs-nats-benchmark-golang/proto"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	mode := flag.String("mode", "grpc", "mode: grpc|nats")
	grpcTarget := flag.String("target", "localhost:50051", "gRPC target")
	natsURL := flag.String("nats_url", "nats://localhost:4222", "NATS URL")
	requests := flag.Int("requests", 50000, "total requests")
	concurrency := flag.Int("concurrency", 200, "concurrency")
	payloadSize := flag.Int("payload_size", 16, "payload size in bytes")
	timeout := flag.Duration("timeout", 500*time.Millisecond, "request timeout")
	flag.Parse()

	payload := make([]byte, *payloadSize)

	// Start NATS subscriber if mode includes NATS
	if *mode == "nats" || *mode == "all" {
		nc, err := nats.Connect(*natsURL, nats.MaxReconnects(10), nats.ReconnectWait(500*time.Millisecond))
		if err != nil {
			log.Fatalf("nats connect: %v", err)
		}
		defer nc.Close()

		// NATS subscriber for "ping" requests
		_, err = nc.Subscribe("ping", func(msg *nats.Msg) {
			msg.Respond(msg.Data) // echo payload
		})
		if err != nil {
			log.Fatalf("nats subscribe: %v", err)
		}

		nc.Flush()
		log.Println("NATS subscriber started for ping requests")
	}

	switch *mode {
	case "grpc":
		conn, err := grpc.Dial(*grpcTarget, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("grpc dial: %v", err)
		}
		defer conn.Close()
		client := pb.NewPingServiceClient(conn)
		bench.RunGRPC(client, payload, *requests, *concurrency, *timeout)

	case "nats":
		nc, err := nats.Connect(*natsURL)
		if err != nil {
			log.Fatalf("nats connect: %v", err)
		}
		defer nc.Close()
		bench.RunNATS(nc, "ping", payload, *requests, *concurrency, *timeout)

	case "all":
		// Run both benchmarks sequentially
		conn, err := grpc.Dial(*grpcTarget, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("grpc dial: %v", err)
		}
		defer conn.Close()
		client := pb.NewPingServiceClient(conn)
		bench.RunGRPC(client, payload, *requests, *concurrency, *timeout)

		nc, err := nats.Connect(*natsURL)
		if err != nil {
			log.Fatalf("nats connect: %v", err)
		}
		defer nc.Close()
		bench.RunNATS(nc, "ping", payload, *requests, *concurrency, *timeout)

	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %s", *mode)
		os.Exit(2)
	}
}
