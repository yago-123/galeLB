package main

import (
	"fmt"
	"log"

	"github.com/yago-123/galelb/pkg/lbnetwork/nodemanager"

	"github.com/sirupsen/logrus"
	lbConfig "github.com/yago-123/galelb/config/lb"
	"github.com/yago-123/galelb/pkg/ring"
	"github.com/yago-123/galelb/pkg/routing"
)

var cfg *lbConfig.Config

func main() {
	// Execute the root command
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting load balancer with config: %v", cfg)

	// Create consistent hashing with 5 virtual nodes per real node
	ch, err := ring.New(ring.Crc32Hasher, 5)
	if err != nil {
		log.Fatalf("Error creating ring: %s", err)
	}

	// Load dummy eBPF program
	router := routing.New(cfg.Logger, "eno1", 8080)
	if errLoad := router.LoadRouter(); errLoad != nil {
		log.Fatalf("failed to load router, ensure you have the required permissions: %s", errLoad)
	}

	// Create gRPC server for managing nodes
	server := nodemanager.New(cfg, 50051)
	server.Start()

	// Add some nodes
	ch.AddNode("Node1")
	ch.AddNode("Node2")
	ch.AddNode("Node3")

	// Hash the IP of a request
	for i := 0; i < 15; i++ {
		cfg.Logger.Infof("request from IP will be routed to %s\n", ch.GetNode([]byte(fmt.Sprintf("113.168.1.1%d", i))))
	}
}
