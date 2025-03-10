package e2e_test

import (
	"context"
	"testing"
	"time"
)

const (
	DefaultIncomingClientRequests = 8080

	SetupReadyStateTimeout  = 15 * time.Second
	VerifyReadyStateTimeout = 10 * time.Second

	SetupStoppedStateTimeout = 10 * time.Second
)

var nodeHosts = []string{ //nolint:gochecknoglobals // OK to have global test data
	"node-0.local",
	"node-1.local",
	"node-2.local",
}

var lbHosts = []string{ //nolint:gochecknoglobals // OK to have global test data
	"lb-0.local",
	"lb-1.local",
	"lb-2.local",
}

var allHosts = append(nodeHosts, lbHosts...) //nolint:gochecknoglobals // OK to have global test data

func setupReadyState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), SetupReadyStateTimeout)
	defer cancel()

	if err := PingHosts(ctx, allHosts); err != nil {
		t.Fatal(err)
	}

	// todo(): set all nodes into ready state
	// todo(): sleep for a while to allow the nodes to be ready
	// todo(): check that load balancer reflects correct changes
}

func verifyReadyState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), VerifyReadyStateTimeout)
	defer cancel()

	if err := PingHosts(ctx, allHosts); err != nil {
		t.Fatal(err)
	}

	if err := CheckNodesServeRequests(ctx, nodeHosts); err != nil {
		t.Fatal(err)
	}

	// if err := CheckLBsForwardRequests(ctx, lbHosts); err != nil {
	// 	t.Fatal(err)
	// }
}

func setupStoppedState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), SetupStoppedStateTimeout)
	defer cancel()

	if err := PingHosts(ctx, allHosts); err != nil {
		t.Fatal(err)
	}
	// todo(): set all nodes into stopped state
	// todo(): sleep for a while to allow the nodes to be stopped
	// todo(): check that load balancer reflects correct changes
}

func TestLoadBalancer_setup(t *testing.T) {
	setupReadyState(t)
	verifyReadyState(t)

}

func TestLoadBalancer_testAllInstancesUpAndRunning(_ *testing.T) {

}
