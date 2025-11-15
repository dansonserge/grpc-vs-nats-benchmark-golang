package nats_server

import (
	"log"

	"github.com/nats-io/nats.go"
)

func Run(nc *nats.Conn, subject string) {
	_, err := nc.QueueSubscribe(subject, "ping-workers", func(msg *nats.Msg) {
		// simply respond with same payload
		if err := msg.Respond(msg.Data); err != nil {
			log.Printf("respond error: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("subscribe error: %v", err)
	}

	// ensure messages are processed
	nc.Flush()
	if err := nc.LastError(); err != nil {
		log.Fatalf("nats error: %v", err)
	}
}
