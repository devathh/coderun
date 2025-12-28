package httpserver

import (
	"context"
	"net"
	"net/http"

	"github.com/devathh/coderun/rest-gateway/internal/infrastructure/config"
)

type Server struct {
	server *http.Server
}

func New(cfg *config.Config, handler http.Handler) *Server {
	addr := net.JoinHostPort(
		cfg.Server.HTTP.Host,
		cfg.Server.HTTP.Port,
	)

	return &Server{
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
