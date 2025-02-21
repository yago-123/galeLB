package lb_net

import (
	"context"

	"github.com/sirupsen/logrus"
	pb "github.com/yago-123/galelb/pkg/consensus"
)

type LBNodeManager struct {
	pb.UnimplementedLBNodeManagerServer

	logger *logrus.Logger
}

func NewNodeManager(logger *logrus.Logger) *LBNodeManager {
	return &LBNodeManager{logger: logger}
}

func (s *LBNodeManager) RegisterNode(ctx context.Context, req *pb.NodeInfo) (*pb.RegisterResponse, error) {
	s.logger.Debugf("Registering node: %s at %s:%d", req.NodeId, req.Ip, req.Port)

	return &pb.RegisterResponse{Success: true, Message: "Node registered successfully"}, nil
}
