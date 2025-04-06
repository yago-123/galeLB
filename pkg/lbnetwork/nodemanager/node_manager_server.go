package nodemanager

import (
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/yago-123/galelb/pkg/util"

	"github.com/yago-123/galelb/pkg/registry"

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
	KeepAliveProbeFrequency = 10 * time.Second
	KeepAliveProbeTimeout   = 15 * time.Second

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
	ip, err := util.GetIPv4FromInterface(s.cfg.PrivateInterface.NetIfacePrivate)
	if err != nil {
		log.Fatalf("Failed to get IPv4 from docker network interface: %v", err)
	}

	listener, err := net.Listen(DefaultL4Protocol, fmt.Sprintf("%s:%d", ip, s.cfg.PrivateInterface.NodePort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if errServe := s.grpcNodesServer.Serve(listener); errServe != nil {
		log.Fatalf("Failed to serve: %v", errServe)
	}
}

func (s *Server) Stop() {

}
