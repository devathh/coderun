package grpcserver

import (
	"fmt"
	"net"

	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/config"
	"google.golang.org/grpc"
)

type Server struct {
	cfg *config.Config
	srv *grpc.Server
}

func New(cfg *config.Config, srv *grpc.Server) *Server {
	return &Server{
		cfg: cfg,
		srv: srv,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen(
		s.cfg.Server.GRPC.Protocol,
		net.JoinHostPort(
			s.cfg.Server.GRPC.Host,
			s.cfg.Server.GRPC.Port,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	if err := s.srv.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve listener: %w", err)
	}

	return nil
}

func (s *Server) GracefulShutdown() {
	s.srv.GracefulStop()
}
