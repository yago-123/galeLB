package nodemanager

import (
	"fmt"
	"github.com/yago-123/galelb/pkg/registry"
	"log"
	"math"
	"net"
	"time"

	lbConfig "github.com/yago-123/galelb/config/lb"

	pb "github.com/yago-123/galelb/pkg/consensus/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	MaxRecvMsgSize = 4 * 1024 * 1024 // 4MB
	MaxSendMsgSize = 4 * 1024 * 1024 // 4MB

	gRPCInfinity            = time.Duration(math.MaxInt64)
	KeepAliveProbeFrequency = time.Second * 3
	KeepAliveProbeTimeout   = time.Second * 5

	DefaultL4Protocol = "tcp"
)

type Server struct {
	grpcNodesServer *grpc.Server

	cfg *lbConfig.Config
}

func New(cfg *lbConfig.Config, registry *registry.NodeRegistry) *Server {
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(MaxRecvMsgSize),
		grpc.MaxSendMsgSize(MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     gRPCInfinity,
			MaxConnectionAge:      gRPCInfinity,
			MaxConnectionAgeGrace: gRPCInfinity,
			Time:                  KeepAliveProbeFrequency,
			Timeout:               KeepAliveProbeTimeout,
		}),
		// grpc.Creds(),                            // todo
		// grpc.UnaryInterceptor(UnaryInterceptor), // todo
		// grpc.StreamInterceptor(),                // todo
	)

	pb.RegisterLBNodeManagerServer(grpcServer, NewNodeManager(cfg, registry))

	// todo() remove once the project has been stabilized
	reflection.Register(grpcServer)

	return &Server{
		grpcNodesServer: grpcServer,
		cfg:             cfg,
	}
}

func (s *Server) Start() {
	listener, err := net.Listen(DefaultL4Protocol, fmt.Sprintf(":%d", s.cfg.Local.NodePort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if errServe := s.grpcNodesServer.Serve(listener); errServe != nil {
		log.Fatalf("Failed to serve: %v", errServe)
	}
}

func (s *Server) Stop() {

}
