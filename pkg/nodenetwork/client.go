package nodenetwork

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client v1Consensus.LBNodeManagerClient

	healthStream grpc.BidiStreamingClient[v1Consensus.HealthStatus, v1Consensus.HealthStatus]

	logger *logrus.Logger
}

func NewClient(logger *logrus.Logger, ip string, port int) (*Client, error) {
	remoteServer := fmt.Sprintf("%s:%d", ip, port)

	// todo(): we must have an array of remove servers for multi-node load balancer
	conn, err := grpc.NewClient(remoteServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not connect to load balancer: %v", err)
	}

	client := v1Consensus.NewLBNodeManagerClient(conn)
	healthStream, err := client.ReportHealthStatus(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not report health status: %v", err)
	}

	return &Client{
		conn:         conn,
		client:       client,
		healthStream: healthStream,
		logger:       logger,
	}, nil
}

func (c *Client) GetConfig(ctx context.Context) (v1Consensus.ConfigResponse, error) {
	config, err := c.client.GetConfig(ctx, &emptypb.Empty{})
	if err != nil {
		c.logger.Errorf("failed to get config: %v", err)
	}

	if config == nil {
		// This should never happen, adding here just in case to avoid panic
		return v1Consensus.ConfigResponse{}, fmt.Errorf("value retrieved in config is nil")
	}

	c.logger.Debugf("received config: %v", config)

	return *config, nil
}

func (c *Client) ReportHealthStatus(ctx context.Context, healthStatus *v1Consensus.HealthStatus) error {
	return c.healthStream.Send(healthStatus)
}
