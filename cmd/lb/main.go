package main

import (
	lbAPIV1 "github.com/yago-123/galelb/pkg/lbnetwork/api/v1"
	"github.com/yago-123/galelb/pkg/lbnetwork/nodemanager"
	"github.com/yago-123/galelb/pkg/registry"

	"github.com/sirupsen/logrus"
	lbConfig "github.com/yago-123/galelb/config/lb"
)

var cfg *lbConfig.Config

func main() {
	// Execute the root command
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting load balancer with config: %v", cfg)

	// Create routing mechanism with consistent hashing (5 virtual nodes per real node)
	// router, err := routing.New(cfg, 5)
	// if err != nil {
	// 	cfg.Logger.Fatalf("failed to create router: %s", err)
	// }

	// Create registry for managing nodes
	nodeRegistry := registry.New(cfg.Logger)

	// Create API for querying load balancer
	lbAPI := lbAPIV1.New(cfg, nodeRegistry)

	// Create gRPC server for managing nodes
	server := nodemanager.New(cfg, nodeRegistry)
	server.Start()

	// Start the load balancer API
	go func() {
		errAPI := lbAPI.Start()
		if errAPI != nil {
			cfg.Logger.Errorf("failed to start load balancer API: %v", errAPI)
		}
	}()
	defer lbAPI.Stop()

	// Add some nodes
	// router.AddNode(common.AddrKey{}, "192.168.1.2", 9091)
	// router.AddNode(common.AddrKey{}, "192.168.1.3", 9091)
	// router.AddNode(common.AddrKey{}, "192.168.1.4", 9091)

	// Hash the IP of a request
	// for i := 0; i < 15; i++ {
	// 	cfg.Logger.Infof("request from IP will be routed to %s\n", router.GetNode([]byte(fmt.Sprintf("113.168.1.1%d", i))))
	// }
}
