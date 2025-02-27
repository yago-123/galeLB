package main

import (
	"context"
	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"
	"time"

	nodeConfig "github.com/yago-123/galelb/config/node"
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
		client, err := nodenetwork.NewClient(cfg.Logger, address.IP, address.Port)
		if err != nil {
			cfg.Logger.Fatalf("failed to create client: %v", err)
		}

		ctxConfig, cancelConfig := context.WithTimeout(context.Background(), ContextTimeout)
		defer cancelConfig()

		executionCfg, err := client.GetConfig(ctxConfig)
		if err != nil {
			cfg.Logger.Fatalf("failed to get config: %v", err)
		}

		normalizedTime := time.Duration(executionCfg.HealthCheckTimeout) * time.Nanosecond
		// todo(): change this to a more accurate value
		healthPeriod := normalizedTime / 2
		for {
			<-time.After(healthPeriod)

			ctx, cancel := context.WithTimeout(context.Background(), normalizedTime)
			defer cancel()

			if err = client.ReportHealthStatus(ctx, &v1Consensus.HealthStatus{
				Service: "gale-node",
				Status:  uint32(v1Consensus.Serving),
				Message: "Serving requests goes brrrrr",
			}); err != nil {
				cfg.Logger.Errorf("failed to report health status: %v", err)
			}
		}
	}
}
