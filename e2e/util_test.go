package e2e_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/yago-123/galelb/pkg/util"
)

// PingHosts pings all the allHosts in the cluster to check if they are reachable and ready for e2e tests
func PingHosts(ctx context.Context, addrs []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(addrs)) // Buffer the channel to hold all potential errors

	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			if err := util.Ping(ctx, addr); err != nil {
				errChan <- err
			}
		}(addr)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check if any error was sent to the error channel
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckNodeServeRequests checks if the node is serving requests correctly by checking if the response body contains
// the host name. In the e2e machine we provisioned a small web server that returns the host name in the response body.
// If the host name is not found in the response body, it means that the node is not serving requests correctly.
func CheckNodeServeRequests(ctx context.Context, host string) error {
	return checkRequest(ctx, host, func(host, body string) bool {
		return strings.Contains(body, host)
	})
}

func CheckNodesServeRequests(ctx context.Context, hosts []string) error {
	for _, host := range hosts {
		if err := CheckNodeServeRequests(ctx, host); err != nil {
			return err
		}
	}

	return nil
}

// CheckLBForwardRequests checks if the load balancer is forwarding requests correctly by checking if the response body
// does not contain the host name. In the e2e machine we provisioned a small web server that returns the host name in
// the response body. If the host name is found in the response body, it means that the load balancer is not forwarding
// the requests
func CheckLBForwardRequests(ctx context.Context, host string) error {
	return checkRequest(ctx, host, func(host, body string) bool {
		return !strings.Contains(body, host)
	})
}

func CheckLBsForwardRequests(ctx context.Context, hosts []string) error {
	for _, host := range hosts {
		if err := CheckLBForwardRequests(ctx, host); err != nil {
			return err
		}
	}

	return nil
}

// checkRequest sends a request to the provided host and checks if the response body is the expected one based
// on the function passed as checkFunc
func checkRequest(ctx context.Context, host string, checkFunc func(string, string) bool) error {
	endpoint := fmt.Sprintf("http://%s", net.JoinHostPort(host, fmt.Sprintf("%d", DefaultIncomingClientRequests)))

	// Create a new request with the provided context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", host, err)
	}

	// Perform the request using the default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to %s: %w", host, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get 200 OK from %s: %d", host, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from %s: %w", host, err)
	}

	// Use the passed checkFunc to determine if the response body is valid for the node or LB
	if !checkFunc(host, string(body)) {
		return fmt.Errorf("host %s failed check, current body: %s", host, string(body))
	}

	return nil
}

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

func verifyStoppedState(_ *testing.T) {

}
