package lb_net

import (
	"log"
	"net"

	"github.com/sirupsen/logrus"
	pb "github.com/yago-123/galelb/pkg/consensus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
}

func New(logger *logrus.Logger) *Server {
	grpcServer := grpc.NewServer()
	pb.RegisterLBNodeManagerServer(grpcServer, NewNodeManager(logger))

	reflection.Register(grpcServer)

	return &Server{grpcServer: grpcServer}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if errServe := s.grpcServer.Serve(listener); errServe != nil {
		log.Fatalf("Failed to serve: %v", errServe)
	}
}

func (s *Server) Stop() {

}
