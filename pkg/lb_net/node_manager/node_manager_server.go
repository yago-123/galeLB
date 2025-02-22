package node_manager

import (
	"fmt"
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
	MaxRecvMsgSize = 1024 * 1024 * 4 // 4MB
	MaxSendMsgSize = 1024 * 1024 * 4 // 4MB

	gRPCInfinity            = time.Duration(math.MaxInt64)
	KeepAliveProbeFrequency = time.Second * 3
	KeepAliveProbeTimeout   = time.Second * 5

	DefaultL4Protocol = "tcp"
)

type Server struct {
	grpcServer *grpc.Server
	port       int

	cfg *lbConfig.Config
}

func New(cfg *lbConfig.Config, port int) *Server {
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

	pb.RegisterLBNodeManagerServer(grpcServer, newNodeManager(cfg))

	// todo() remove once the project has been stabilized
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		port:       port,
	}
}

func (s *Server) Start() {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	listener, err := net.Listen(DefaultL4Protocol, addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if errServe := s.grpcServer.Serve(listener); errServe != nil {
		log.Fatalf("Failed to serve: %v", errServe)
	}
}

func (s *Server) Stop() {

}
