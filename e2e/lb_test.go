package e2e_test

import (
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

func TestLoadBalancer_setup(t *testing.T) {
	setupReadyState(t)
	verifyReadyState(t)
	setupStoppedState(t)
	verifyStoppedState(t)
}

func TestLoadBalancer_testAllInstancesUpAndRunning(_ *testing.T) {

}
