package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yago-123/galelb/pkg/node_net"
)

const (
	ContextTimeout = time.Second * 5
)

var logger = logrus.New()

func main() {
	logger.SetLevel(logrus.DebugLevel)

	client := node_net.New(logger, "192.168.18.130", 50051)

	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
	defer cancel()

	if err := client.RegisterNode(ctx); err != nil {
		logger.Errorf("failed to register node: %v", err)
	}
}
