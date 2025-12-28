package rediscache

import (
	"context"
	"fmt"
	"net"

	"github.com/devathh/coderun/sso-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg *config.Config) (*redis.Client, error) {
	if cfg == nil {
		return nil, customerrors.ErrNilArgs
	}

	client := redis.NewClient(&redis.Options{
		Addr: net.JoinHostPort(
			cfg.Secrets.Redis.Host,
			cfg.Secrets.Redis.Port,
		),
		Password: cfg.Secrets.Redis.Password,
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

func Close(client *redis.Client) error {
	return client.Conn().Close()
}
