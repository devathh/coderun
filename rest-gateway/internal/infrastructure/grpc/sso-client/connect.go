package ssoclient

import (
	"fmt"
	"net"

	ssopb "github.com/devathh/coderun/rest-gateway/api/sso/v1"
	"github.com/devathh/coderun/rest-gateway/internal/infrastructure/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect(cfg *config.Config) (ssopb.SSOClient, *grpc.ClientConn, error) {
	addr := net.JoinHostPort(
		cfg.Services.CoderunSSO.Host,
		cfg.Services.CoderunSSO.Port,
	)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to sso: %w", err)
	}

	client := ssopb.NewSSOClient(conn)

	return client, conn, nil
}

func Close(conn *grpc.ClientConn) error {
	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection with sso: %w", err)
	}

	return nil
}
