package main

import (
	"context"
	"fmt"
	"net"
	"time"

	nodeAPIV1 "github.com/yago-123/galelb/pkg/nodenetwork/api/v1"

	"github.com/yago-123/galelb/pkg/util"

	nodeConfig "github.com/yago-123/galelb/config/node"
	nodeNet "github.com/yago-123/galelb/pkg/nodenetwork"

	"github.com/sirupsen/logrus"
)

const (
	HostnameResolveTimeout = 5 * time.Second
)

var cfg *nodeConfig.Config

func main() {
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting node with config: %v", cfg)

	targets, err := retrieveIPAndPorts(cfg)
	if err != nil {
		cfg.Logger.Fatalf("failed to retrieve IP and ports: %v", err)
	}

	// Create dispatcher for managing requests towards the load balancers
	dispatcher := nodeNet.NewDispatcher(cfg, targets)

	// Create API for querying the node
	nodeApi := nodeAPIV1.New(cfg, dispatcher)

	nodeApi.Start()
	defer nodeApi.Stop()

	dispatcher.Start()
	defer dispatcher.Stop()

	// todo(); add some logic for stopping the node in here
}

// retrieveIPAndPorts retrieves the IP addresses and ports from the configuration file and returns a map of targets. It
// also does basic checking and hostname resolution if the IP is not set.
func retrieveIPAndPorts(cfg *nodeConfig.Config) (map[string]nodeNet.Target, error) {
	var err error
	targets := make(map[string]nodeNet.Target)

	for idx, address := range cfg.LoadBalancer.Addresses {
		if address.Port == 0 {
			return nil, fmt.Errorf("port at config index %d is not set", idx)
		}

		if address.IP == "" && address.Hostname == "" {
			return nil, fmt.Errorf("invalid address configuration, IP nor hostname is defined for index %d", idx)
		}

		// If the IP is not set, resolve the hostname via multicast DNS or regular DNS
		if address.IP == "" && address.Hostname != "" {
			ips := []net.IP{}

			ctx, cancel := context.WithTimeout(context.Background(), HostnameResolveTimeout)
			defer cancel()

			if util.IsMultiCastDNS(address.Hostname) {
				ips, err = util.ResolveMulticastDNS(ctx, address.Hostname)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve multicast DNS %s: %v", address.Hostname, err)
				}
			} else if !util.IsMultiCastDNS(address.Hostname) {
				ips, err = util.ResolveDNS(ctx, address.Hostname, "127.0.0.1:53")
				if err != nil {
					return nil, fmt.Errorf("failed to resolve hostname from configuration %s: %v", address.Hostname, err)
				}
			}

			if len(ips) == 0 {
				return nil, fmt.Errorf("no IP addresses found for hostname: %s", address.Hostname)
			}

			cfg.Logger.Debugf("resolved hostname %s to IP %s", address.Hostname, ips[0].String())

			// notice this will not overwrite the original configuration
			address.IP = ips[0].String()
		}

		target := nodeNet.Target{
			IP:   address.IP,
			Port: address.Port,
		}

		targets[target.String()] = target
	}

	return targets, nil
}
