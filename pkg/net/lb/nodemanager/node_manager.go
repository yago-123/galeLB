package nodemanager

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	nodeID string
}

// nodeManager contains the logic for synchronizing the load balancer with the nodeRegistry. It's structured as a reverse
// console, allowing nodeRegistry to connect to the server as soon as they start in order to be registered and allocated
// traffic
type nodeManager struct {
	cfg              *lbConfig.Config
	nodeRegistry     map[string]node
	nodeRegistryLock sync.RWMutex

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

// RegisterNode ensures that the new node spawned is accepted into the load balancer registry
func (s *nodeManager) RegisterNode(_ context.Context, req *v1Consensus.NodeInfo) (*v1Consensus.RegisterResponse, error) {
	s.logger.Debugf("Registering node: %s at %s:%d", req.GetNodeId(), req.GetIp(), req.GetPort())

	// todo: append nodeRegistry
	if _, ok := s.nodeRegistry[req.GetNodeId()]; ok {
		s.logger.Debugf("tried to register node %s with addrs %s:%d already registered", req.GetNodeId(), req.GetIp(), req.GetPort())
		return &v1Consensus.RegisterResponse{Success: false, Message: "Node was already present in the registry"}, nil
	}

	s.logger.Debugf("registered new node %s with addrs: %s:%d", req.GetNodeId(), req.GetIp(), req.GetPort())
	s.nodeRegistry[req.GetNodeId()] = node{nodeID: req.GetNodeId()}

	return &v1Consensus.RegisterResponse{Success: true, Message: "Node registered successfully"}, nil
}

func (s *nodeManager) ReportHealthStatus(stream grpc.BidiStreamingServer[v1Consensus.HealthStatus, v1Consensus.HealthStatus]) error {
	msgChan := make(chan *v1Consensus.HealthStatus, ChannelBufferSize)
	errChan := make(chan error, ChannelBufferSize)

	defer close(msgChan)
	defer close(errChan)

	// Set timeout for health checks. Nodes should send health checks at least once every half of this duration
	timer := time.NewTimer(s.cfg.NodeHealthChecksTimeout)
	defer timer.Stop()

	go func() {
		for {
			req, err := stream.Recv()
			if gRPCErrUnrecoverable(err) {
				// todo(): add some sort of metadata id to identify the streams?
				s.logger.Infof("stream closed by node")
				errChan <- err
				return
			}

			if err != nil {
				errChan <- err
			}

			msgChan <- req
		}
	}()

	for {
		select {
		case msg := <-msgChan:
			s.logger.Infof("received health status from node %s: %v", msg.GetNodeId(), msg.GetStatus())

			// Drain and reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(s.cfg.NodeHealthChecksTimeout)
		case err := <-errChan:
			s.logger.Errorf("error receiving health status: %v", err)
			if gRPCErrUnrecoverable(err) {
				s.unregisterNode()
				return fmt.Errorf("unrecoverable error receiving health status: %w", err)
			}

			// Do not drain the timer, as we want to stop tracking this node if it does not send health status
			// todo(): may be worth to send (timeout/2) - 1?
		case <-timer.C:
			s.unregisterNode()
			// todo(): update the node info
			return fmt.Errorf("timed out waiting for health status")
		}
	}
}

// unregisterNode removes a node from the registry
func (s *nodeManager) unregisterNode() {

	// todo(): it should also trigger a leader election somehow to re-allocate the traffic
}

// registerNode adds a node to the registry
func (s *nodeManager) registerNode() {

}

func gRPCErrUnrecoverable(err error) bool {
	return status.Code(err) == codes.Canceled || status.Code(err) == codes.Unavailable
}
