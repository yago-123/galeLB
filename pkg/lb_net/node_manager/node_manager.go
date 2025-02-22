package node_manager

import (
	"context"

	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"

	"github.com/sirupsen/logrus"
)

// nodeManager contains the logic for synchronizing the load balancer with the nodes. It's structured as a reverse
// console, allowing nodes to connect to the server as soon as they start in order to be registered and allocated
// traffic
type nodeManager struct {
	v1Consensus.UnimplementedLBNodeManagerServer

	logger *logrus.Logger
}

func newNodeManager(logger *logrus.Logger) *nodeManager {
	return &nodeManager{logger: logger}
}

func (s *nodeManager) RegisterNode(ctx context.Context, req *v1Consensus.NodeInfo) (*v1Consensus.RegisterResponse, error) {
	s.logger.Debugf("Registering node: %s at %s:%d", req.NodeId, req.Ip, req.Port)

	return &v1Consensus.RegisterResponse{Success: true, Message: "Node registered successfully"}, nil
}
