package main

import (
	"context"
	nodeAPIV1 "github.com/yago-123/galelb/pkg/nodenetwork/api/v1"
	"net"
	"sync"
	"time"

	"github.com/yago-123/galelb/pkg/util"

	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"

	nodeConfig "github.com/yago-123/galelb/config/node"
	nodeNet "github.com/yago-123/galelb/pkg/nodenetwork"

	"github.com/sirupsen/logrus"
)

const (
	GetConfigTimeout       = 5 * time.Second
	HostnameResolveTimeout = 5 * time.Second
)

var cfg *nodeConfig.Config

func main() {
	var err error
	var wg sync.WaitGroup

	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting node with config: %v", cfg)

	dispatcher := nodeNet.NewDispatcher()

	nodeAPIV1.New(cfg, dispatcher)

	dispatcher.Start()
	defer dispatcher.Stop()

	for idx, address := range cfg.LoadBalancer.Addresses {
		if address.IP == "" && address.Hostname == "" {
			cfg.Logger.Fatalf("invalid address configuration, IP nor hostname is defined for index %d", idx)
		}

		// If the IP is not set, resolve the hostname via multicast DNS or regular DNS
		if address.IP == "" && address.Hostname != "" {
			ips := []net.IP{}

			ctx, cancel := context.WithTimeout(context.Background(), HostnameResolveTimeout)
			defer cancel()

			if util.IsMultiCastDNS(address.Hostname) {
				ips, err = util.ResolveMulticastDNS(ctx, address.Hostname)
				if err != nil {
					cfg.Logger.Warnf("failed to resolve multicast DNS: %v", err)
				}
			} else if !util.IsMultiCastDNS(address.Hostname) {
				ips, err = util.ResolveDNS(ctx, address.Hostname, "127.0.0.1:53")
				if err != nil {
					cfg.Logger.Warnf("failed to resolve hostname from configuration %s: %v", address.Hostname, err)
				}
			}

			if len(ips) == 0 {
				cfg.Logger.Fatalf("no IP addresses found for hostname: %s", address.Hostname)
			}

			cfg.Logger.Debugf("resolved hostname %s to IP %s", address.Hostname, ips[0].String())
			// todo(): this will not overwrite the original configuration
			address.IP = ips[0].String()
		}

		client, err := nodeNet.NewClient(cfg.Logger, address.IP, address.Port)
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

		// Increase work group to wait for the health status goroutine and spawn health check reporter
		wg.Add(1)
		go func() {
			// Ensure Done is called when the goroutine finishes
			defer wg.Done()

			// Report health status to the load balancer
			for {
				ctx, cancel := context.WithTimeout(context.Background(), normalizedTime)
				defer cancel()

				if err = client.ReportHealthStatus(ctx, &v1Consensus.HealthStatus{
					Service: "gale-node",
					Status:  uint32(v1Consensus.Serving),
					Message: "Serving requests goes brrrrr",
				}); err != nil {
					cfg.Logger.Errorf("failed to report health status: %v", err)
				}

				cfg.Logger.Debugf("reported health status to %s:%d", address.IP, address.Port)

				<-time.After(healthPeriod)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
