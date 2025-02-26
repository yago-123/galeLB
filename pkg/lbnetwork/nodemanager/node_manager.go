package nodemanager

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"sync"
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
}

// nodeManager contains the logic for synchronizing the load balancer with the nodeRegistry. It's structured as a reverse
// console, allowing nodeRegistry to connect to the server as soon as they start in order to be registered and allocated
// traffic
type nodeManager struct {
	cfg                   *lbConfig.Config
	nodeRegistry          map[string]node
	nodeRegistryBlackList map[string]time.Time
	nodeRegistryLock      sync.RWMutex

	// internal structure required for gRPC implementation
	v1Consensus.UnimplementedLBNodeManagerServer

	logger *logrus.Logger
}

func newNodeManager(cfg *lbConfig.Config) *nodeManager {
	return &nodeManager{
		cfg:          cfg,
		nodeRegistry: map[string]node{},
		logger:       cfg.Logger,
	}
}

func (s *nodeManager) GetConfig(ctx context.Context, req *emptypb.Empty) (*v1Consensus.ConfigResponse, error) {
	// todo(): return s.cfg filtered to only the necessary fields
	return &v1Consensus.ConfigResponse{}, nil
}

// ListenLoop is the main loop for listening to health checks from nodes. Nodes send health checks periodically to the
// LB to indicate their presence. If a node does not send a health check within a certain timeout, it is removed.
func (s *nodeManager) ListenLoop(stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) error {
	msgChan := make(chan *v1Consensus.HealthStatus, ChannelBufferSize)
	errChan := make(chan error, ChannelBufferSize)

	defer close(msgChan)
	defer close(errChan)

	tcpAddr, err := extractPeerInfoFromConn(stream)
	if err != nil {
		return fmt.Errorf("failed to extract peer info from stream: %w", err)
	}

	// nodeKey will be used to access the node registry-related info for the node
	nodeKey := tcpAddr.String()

	s.logger.Debugf("registered new health check probe from %s", tcpAddr.String())

	// Set timeout for health checks. Nodes should send health checks at least once every half of this duration
	timer := time.NewTimer(s.cfg.NodeHealth.ChecksTimeout)
	defer timer.Stop()

	// Spawn async function for listening for health checks from nodes
	go func() {
		for {
			// Wait for new updates from the node
			req, errRecv := stream.Recv()
			if gRPCErrUnrecoverable(errRecv) {
				s.logger.Infof("stream closed by node %s", nodeKey)
				errChan <- errRecv
				return
			}

			if err != nil {
				errChan <- errRecv
			}

			msgChan <- req
		}
	}()

	// Main loop for receiving health checks
	for {
		select {
		case _ = <-msgChan:
			s.logger.Infof("received health check from node %s", nodeKey)

			if !s.nodeAlreadyPresent(nodeKey) {
				s.registerNode(nodeKey)
			}

			// Drain and reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(s.cfg.NodeHealth.ChecksTimeout)
		case err := <-errChan:
			s.logger.Errorf("error receiving health status: %v", err)
			if gRPCErrUnrecoverable(err) {
				s.unregisterNode(nodeKey)
				return fmt.Errorf("unrecoverable error receiving health status: %w", err)
			}

			// Do not drain the timer, as we want to stop tracking this node if it does not send health status
			// todo(): may be worth to send (timeout/2) - 1?
		case <-timer.C:
			s.unregisterNode(nodeKey)
			return fmt.Errorf("timed out waiting for health status from %s", nodeKey)
		}
	}
}

// unregisterNode removes a node from the registry
func (s *nodeManager) unregisterNode(nodeKey string) {
	s.nodeRegistryLock.Lock()
	defer s.nodeRegistryLock.Unlock()

	// todo(): it should also trigger a leader election somehow to re-allocate the traffic

	// todo(): before deleting check if exists
	if _, ok := s.nodeRegistry[nodeKey]; ok {
		delete(s.nodeRegistry, nodeKey)
	}
}

// registerNode adds a node to the registry
func (s *nodeManager) registerNode(nodeKey string) {
	s.nodeRegistryLock.Lock()
	defer s.nodeRegistryLock.Unlock()

	s.nodeRegistry[nodeKey] = node{}
}

func (s *nodeManager) nodeAlreadyPresent(nodeKey string) bool {
	s.nodeRegistryLock.RLock()
	defer s.nodeRegistryLock.RUnlock()

	// todo(): probably can use specific hasKeys func
	_, ok := s.nodeRegistry[nodeKey]
	return ok
}

func gRPCErrUnrecoverable(err error) bool {
	return status.Code(err) == codes.Canceled || status.Code(err) == codes.Unavailable
}

func extractPeerInfoFromConn(stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) (net.TCPAddr, error) {
	p, ok := peer.FromContext(stream.Context())
	if ok {
		if addr, okTcp := p.Addr.(*net.TCPAddr); okTcp {
			return *addr, nil
		}
	}

	return net.TCPAddr{}, fmt.Errorf("failed to extract peer info from stream")
}
