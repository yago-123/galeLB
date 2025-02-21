package main

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/yago-123/galelb/pkg/lb_net"
	"github.com/yago-123/galelb/pkg/ring"
	"github.com/yago-123/galelb/pkg/routing"
)

var logger = logrus.New()

func main() {
	logger.SetLevel(logrus.DebugLevel)

	// Create consistent hashing with 5 virtual nodes per real node
	ch, err := ring.New(ring.Crc32Hasher, 5)
	if err != nil {
		log.Fatalf("Error creating ring: %s", err)
	}

	// Load dummy eBPF program
	router := routing.New(logger, "eno1")
	if err := router.LoadRouter(); err != nil {
		log.Fatalf("failed to load router: %s", err)
	}

	// create gRPC server for managing nodes
	server := lb_net.New(logger)

	server.Start()

	// Add some nodes
	ch.AddNode("Node1")
	ch.AddNode("Node2")
	ch.AddNode("Node3")

	// Hash the IP of a request
	for i := 0; i < 15; i++ {
		fmt.Printf("Request from IP will be routed to %s\n", ch.GetNode([]byte(fmt.Sprintf("113.168.1.1%d", i))))
	}
}
