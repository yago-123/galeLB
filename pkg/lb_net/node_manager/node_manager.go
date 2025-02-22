package node_manager

import (
	"context"
	"fmt"
	lbConfig "github.com/yago-123/galelb/config/lb"
	"google.golang.org/grpc"
	"sync"
	"time"

	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"

	"github.com/sirupsen/logrus"
)

type node struct {
	nodeID string
}

// nodeManager contains the logic for synchronizing the load balancer with the nodeRegistry. It's structured as a reverse
// console, allowing nodeRegistry to connect to the server as soon as they start in order to be registered and allocated
// traffic
type nodeManager struct {
	v1Consensus.UnimplementedLBNodeManagerServer

	nodeRegistry     map[string]node
	nodeRegistryLock sync.RWMutex

	logger *logrus.Logger
}

func newNodeManager(cfg *lbConfig.Config) *nodeManager {
	return &nodeManager{
		nodeRegistry: map[string]node{},
		logger:       cfg.Logger,
	}
}

// RegisterNode ensures that the new node spawned is accepted into the load balancer registry
func (s *nodeManager) RegisterNode(ctx context.Context, req *v1Consensus.NodeInfo) (*v1Consensus.RegisterResponse, error) {
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
	msgChan := make(chan *v1Consensus.HealthStatus)
	errChan := make(chan error)

	// set
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	for {
		go func() {
			for {
				req, err := stream.Recv()
				if err != nil {
					errChan <- err
				}

				msgChan <- req
			}
		}()

		for {
			select {
			//case msg := <-msgChan:
			// case err := <-errChan:
			case <-timer.C:
				return fmt.Errorf("timed out waiting for health status")
			}
		}
	}

	return nil
}
