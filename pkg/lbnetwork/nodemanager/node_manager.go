package nodemanager

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/yago-123/galelb/pkg/registry"

	"github.com/yago-123/galelb/pkg/util"

	"google.golang.org/protobuf/types/known/emptypb"

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

// NodeManager contains the logic for synchronizing the load balancer with the NodeRegistry. It's structured as a reverse
// console, allowing NodeRegistry to connect to the server as soon as they start in order to be registered and allocated
// traffic
type NodeManager struct {
	cfg *lbConfig.Config

	// registry is the internal structure that keeps track of the nodes and their health status
	registry *registry.NodeRegistry

	//

	// Internal structure required for gRPC implementation
	v1Consensus.UnimplementedLBNodeManagerServer

	logger *logrus.Logger
}

func NewNodeManager(cfg *lbConfig.Config, registry *registry.NodeRegistry) *NodeManager {
	return &NodeManager{
		cfg:      cfg,
		registry: registry,
		logger:   cfg.Logger,
	}
}

// GetConfig returns the current configuration of the load balancer so that nodes can adjust their parameters accordingly
func (s *NodeManager) GetConfig(_ context.Context, _ *emptypb.Empty) (*v1Consensus.ConfigResponse, error) {
	return &v1Consensus.ConfigResponse{
		ChecksBeforeRouting: uint32(s.cfg.NodeHealth.ChecksBeforeRouting), //nolint:gosec // secure to do this uint conversion
		HealthCheckTimeout:  s.cfg.NodeHealth.ChecksTimeout.Nanoseconds(),
		BlackListAfterFails: int64(s.cfg.NodeHealth.BlackListAfterFails),
		BlackListExpiry:     s.cfg.NodeHealth.BlackListExpiry.Nanoseconds(),
	}, nil
}

// ReportHealthStatus is the main loop for listening to health checks from nodes. Nodes send health checks periodically to the
// LB to indicate their presence. If a node does not send a health check within a certain timeout, it is removed.
func (s *NodeManager) ReportHealthStatus(stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) error {
	msgChan := make(chan *v1Consensus.HealthStatus, ChannelBufferSize)
	errChan := make(chan error, ChannelBufferSize)

	defer close(msgChan)
	defer close(errChan)

	// nodeKey will be used to access the node registry-related info for the node
	tcpAddr, err := extractTCPFromConn(stream)
	if err != nil {
		return fmt.Errorf("failed to extract peer info from stream: %w", err)
	}

	// Try to retrieve the MAC address from the ARP cache. If it fails, try to get it via an ARP call
	mac, err := util.GetMACFromARPCache(tcpAddr.IP.String(), s.cfg.Local.NetIfaceNodes)
	if err != nil {
		s.logger.Warnf("failed to get MAC address from ARP cache: %v", err)

		mac, err = util.GetMACViaARPCall(tcpAddr.IP.String(), s.cfg.Local.NetIfaceNodes)
		if err != nil {
			s.logger.Errorf("failed to get MAC address via ARP call: %v", err)
			return fmt.Errorf("failed to get MAC address via ARP call: %w", err)
		}
	}

	// todo(): replace this
	nodeKey := tcpAddr.String()

	// Register the node if it is not already present
	s.registry.RegisterNode(nodeKey)

	s.logger.Debugf("registered new connection from node %s with mac %s", nodeKey, mac)

	// Spawn async function for listening for health checks from nodes
	go s.listenerReportHealthStatus(nodeKey, msgChan, errChan, stream)

	// Main loop for multiplexing health checks with errors and health check timeouts. Once this function returns it
	// means that there has been an unrecoverable error or the node has been marked as unhealthy
	return s.multiplexHealthStatus(nodeKey, msgChan, errChan)
}

// listenerReportHealthStatus is a helper function for listening to health checks from nodes. It abstracts the listener
// logic from the main function to make the code more readable
func (s *NodeManager) listenerReportHealthStatus(nodeKey string, msgChan chan *v1Consensus.HealthStatus, errChan chan error, stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) {
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

// multiplexHealthStatus is in charge of multiplexing health status updates from nodes. It listens for health status
// updates and errors from the node. If a node does not send a health check within a certain timeout, it is marked as
// unhealthy and the traffic is rerouted to other nodes
func (s *NodeManager) multiplexHealthStatus(nodeKey string, msgChan chan *v1Consensus.HealthStatus, errChan chan error) error {
	// Set timeout for health checks. Nodes should send health checks at least once every half of this duration
	timer := time.NewTimer(s.cfg.NodeHealth.ChecksTimeout)
	defer timer.Stop()

	for {
		select {
		case msg := <-msgChan:
			if msg.GetStatus() == uint32(v1Consensus.NotServing) {
				// todo(): think what to do, we must re-route traffic for sure
				continue
			} else if msg.GetStatus() == uint32(v1Consensus.ShuttingDown) {
				// todo(): invoke quorum and re-route all traffic to other nodes
				s.logger.Infof("node %s is shutting down", nodeKey)
				return nil // todo(): change this return
			}

			// If status is v1Consensus.Serving keep running the loop
			s.registry.ReportNewHealthCheck(nodeKey)

			// Drain and reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(s.cfg.NodeHealth.ChecksTimeout)

		case err := <-errChan:
			s.logger.Errorf("error receiving health status: %v", err)
			if gRPCErrUnrecoverable(err) {
				s.registry.ReportNodeFailure(nodeKey)
				// todo(): trigger action for start rerouting traffic
				// todo(): do we really want to register/unregister or just keep a latest timestamp? s.unregisterNode(nodeKey)
				return fmt.Errorf("unrecoverable error receiving health status: %w", err)
			}

			// Do not drain the timer, as we want to stop tracking this node if it does not send health status
			// todo(): may be worth to send (timeout/2) - 1?
		case <-timer.C:
			s.registry.ReportNodeFailure(nodeKey)
			// todo(): trigger action for start rerouting traffic
			// todo(): do we really want to register/unregister or just keep a latest timestamp? s.unregisterNode(nodeKey) s.unregisterNode(nodeKey)
			return fmt.Errorf("timed out waiting for health status from %s", nodeKey)
		}
	}
}

// gRPCErrUnrecoverable checks if an error is unrecoverable. This is useful for checking if a stream has been closed
// indefinitely or if the connection will be unavailable for a long time
func gRPCErrUnrecoverable(err error) bool {
	return status.Code(err) == codes.Canceled || status.Code(err) == codes.Unavailable
}

// extractTCPFromConn extracts the node key from the connection. Required for uniquely identifying nodes in the
// registry
func extractTCPFromConn(stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) (net.TCPAddr, error) {
	p, ok := peer.FromContext(stream.Context())
	if ok {
		if addr, okTCP := p.Addr.(*net.TCPAddr); okTCP {
			// Make sure that addr is not nil just in case, it should never be nil
			if addr != nil {
				return *addr, nil
			}
		}
	}

	return net.TCPAddr{}, fmt.Errorf("failed to extract peer info from stream")
}
