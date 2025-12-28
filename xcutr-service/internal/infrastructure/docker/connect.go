package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

func Connect() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	if _, err := cli.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping docker: %w", err)
	}

	return cli, nil
}
