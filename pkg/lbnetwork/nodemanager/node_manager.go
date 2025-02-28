package nodemanager

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"time"

	"google.golang.org/grpc/peer"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	lbConfig "github.com/yago-123/galelb/config/lb"
	"google.golang.org/grpc"

	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"

	"github.com/sirupsen/logrus"
)

const (
	ChannelBufferSize = 1
)

type node struct {
	continuousHealthChecks uint
	lastHealthCheck        time.Time
}

// nodeManager contains the logic for synchronizing the load balancer with the nodeRegistry. It's structured as a reverse
// console, allowing nodeRegistry to connect to the server as soon as they start in order to be registered and allocated
// traffic
type nodeManager struct {
	cfg *lbConfig.Config

	registry *nodeRegistry

	// Internal structure required for gRPC implementation
	v1Consensus.UnimplementedLBNodeManagerServer

	logger *logrus.Logger
}

func newNodeManager(cfg *lbConfig.Config) *nodeManager {
	return &nodeManager{
		cfg:      cfg,
		registry: newNodeRegistry(cfg.Logger),
		logger:   cfg.Logger,
	}
}

// GetConfig returns the current configuration of the load balancer so that nodes can adjust their parameters accordingly
func (s *nodeManager) GetConfig(ctx context.Context, _ *emptypb.Empty) (*v1Consensus.ConfigResponse, error) {
	return &v1Consensus.ConfigResponse{
		ChecksBeforeRouting: uint32(s.cfg.NodeHealth.ChecksBeforeRouting),
		HealthCheckTimeout:  s.cfg.NodeHealth.ChecksTimeout.Nanoseconds(),
		BlackListAfterFails: int64(s.cfg.NodeHealth.BlackListAfterFails),
		BlackListExpiry:     s.cfg.NodeHealth.BlackListExpiry.Nanoseconds(),
	}, nil
}

// ReportHealthStatus is the main loop for listening to health checks from nodes. Nodes send health checks periodically to the
// LB to indicate their presence. If a node does not send a health check within a certain timeout, it is removed.
func (s *nodeManager) ReportHealthStatus(stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) error {
	msgChan := make(chan *v1Consensus.HealthStatus, ChannelBufferSize)
	errChan := make(chan error, ChannelBufferSize)

	defer close(msgChan)
	defer close(errChan)

	// nodeKey will be used to access the node registry-related info for the node
	nodeKey, err := extractNodeKeyFromConn(stream)
	if err != nil {
		return fmt.Errorf("failed to extract peer info from stream: %w", err)
	}

	// Register the node if it is not already present
	s.registry.registerNode(nodeKey)

	s.logger.Debugf("registered new connection from node %s", nodeKey)

	// Set timeout for health checks. Nodes should send health checks at least once every half of this duration
	timer := time.NewTimer(s.cfg.NodeHealth.ChecksTimeout)
	defer timer.Stop()

	// Spawn async function for listening for health checks from nodes
	go s.listenerReportHealthStatus(nodeKey, msgChan, errChan, stream)

	// Main loop for multiplexing health checks with errors and health check timeouts
	for {
		select {
		case msg := <-msgChan:
			s.logger.Infof("received health check from node %s with status %d", nodeKey, msg.GetStatus())

			if msg.Status == uint32(v1Consensus.NotServing) {
				// todo(): think what to do, we must re-route traffic for sure
			} else if msg.Status == uint32(v1Consensus.ShuttingDown) {
				// todo(): invoke quorum and re-route all traffic to other nodes
				s.logger.Infof("node %s is shutting down", nodeKey)
				return nil
			}

			// If status is v1Consensus.Serving keep running the loop
			s.registry.reportNewHealthCheck(nodeKey)

			// Drain and reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(s.cfg.NodeHealth.ChecksTimeout)

		case err := <-errChan:
			s.logger.Errorf("error receiving health status: %v", err)
			if gRPCErrUnrecoverable(err) {
				s.registry.reportNodeFailure(nodeKey)
				// todo(): trigger action for start rerouting traffic
				// todo(): do we really want to register/unregister or just keep a latest timestamp? s.unregisterNode(nodeKey)
				return fmt.Errorf("unrecoverable error receiving health status: %w", err)
			}

			// Do not drain the timer, as we want to stop tracking this node if it does not send health status
			// todo(): may be worth to send (timeout/2) - 1?
		case <-timer.C:
			s.registry.reportNodeFailure(nodeKey)
			// todo(): trigger action for start rerouting traffic
			// todo(): do we really want to register/unregister or just keep a latest timestamp? s.unregisterNode(nodeKey) s.unregisterNode(nodeKey)
			return fmt.Errorf("timed out waiting for health status from %s", nodeKey)
		}
	}
}

// listenerReportHealthStatus is a helper function for listening to health checks from nodes. It abstracts the listener
// logic from the main function to make the code more readable
func (s *nodeManager) listenerReportHealthStatus(nodeKey string, msgChan chan *v1Consensus.HealthStatus, errChan chan error, stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) {
	for {
		// Wait for new updates from the node
		req, errRecv := stream.Recv()
		if gRPCErrUnrecoverable(errRecv) {
			s.logger.Infof("stream closed by node %s", nodeKey)
			errChan <- errRecv
			return
		}

		s.logger.Infof("received health check from node %s with status %d", nodeKey, req.GetStatus())

		if errRecv != nil {
			errChan <- errRecv
		}

		msgChan <- req
	}
}

// gRPCErrUnrecoverable checks if an error is unrecoverable. This is useful for checking if a stream has been closed
// indefinitely or if the connection will be unavailable for a long time
func gRPCErrUnrecoverable(err error) bool {
	return status.Code(err) == codes.Canceled || status.Code(err) == codes.Unavailable
}

// extractNodeKeyFromConn extracts the node key from the connection. Required for uniquely identifying nodes in the
// registry
func extractNodeKeyFromConn(stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) (string, error) {
	p, ok := peer.FromContext(stream.Context())
	if ok {
		if addr, okTcp := p.Addr.(*net.TCPAddr); okTcp {
			// Make sure that addr is not nil just in case, it should never be nil
			if addr != nil {
				return addr.String(), nil
			}
		}
	}

	return "", fmt.Errorf("failed to extract peer info from stream")
}
