package main

import (
	"context"
	nodeConfig "github.com/yago-123/galelb/config/node"
	"time"

	"github.com/yago-123/galelb/pkg/nodenetwork"

	"github.com/sirupsen/logrus"
)

const (
	ContextTimeout = time.Second * 5
)

var cfg *nodeConfig.Config

func main() {
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting load balancer with config: %v", cfg)

	for _, address := range cfg.LoadBalancer.Addresses {
		client := nodenetwork.New(cfg.Logger, address.IP, address.Port)

		ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
		defer cancel()

		if err := client.RegisterNode(ctx); err != nil {
			cfg.Logger.Errorf("failed to register node: %v", err)
		}
	}
}
