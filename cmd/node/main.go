package main

import (
	"context"
	"time"

	"github.com/yago-123/galelb/pkg/util"

	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"

	nodeConfig "github.com/yago-123/galelb/config/node"
	"github.com/yago-123/galelb/pkg/nodenetwork"

	"github.com/sirupsen/logrus"
)

const (
	GetConfigTimeout       = time.Second * 5
	HostnameResolveTimeout = 5 * time.Second
)

var cfg *nodeConfig.Config

func main() {
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting node with config: %v", cfg)

	for idx, address := range cfg.LoadBalancer.Addresses {
		if address.IP == "" && address.Hostname == "" {
			cfg.Logger.Fatalf("invalid address configuration, IP nor hostname is defined for index %d", idx)
		}

		// If the IP is not set, resolve the hostname
		if address.IP == "" && address.Hostname != "" {
			ctx, cancel := context.WithTimeout(context.Background(), HostnameResolveTimeout)
			defer cancel()

			ip, err := util.ResolveHostname(ctx, address.Hostname)
			if err != nil {
				cfg.Logger.Fatalf("failed to resolve hostname from configuration %s: %w", address.Hostname, err)
			}

			// Update the address with the resolved IP. Notice this is not saved in the configuration file itself,
			// this is a temporary change
			// todo(): save this change in the configuration file
			address.IP = ip
		}

		client, err := nodenetwork.NewClient(cfg.Logger, address.IP, address.Port)
		if err != nil {
			cfg.Logger.Fatalf("failed to create client: %v", err)
		}

		ctxConfig, cancelConfig := context.WithTimeout(context.Background(), GetConfigTimeout)
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
