package main

import (
	"flag"
	"log"

	"grpc-vs-nats-benchmark-golang/internal/nats_server"

	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := flag.String("nats", nats.DefaultURL, "NATS server URL")
	subject := flag.String("subject", "ping", "NATS subject to subscribe to")
	flag.Parse()

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer nc.Close()

	log.Printf("NATS subscriber listening on %s (subject=%s)", *natsURL, *subject)

	nats_server.Run(nc, *subject)

	// block until killed
	select {}
}
