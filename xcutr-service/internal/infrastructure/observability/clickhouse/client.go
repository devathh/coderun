package clickhouse

import (
	"context"
	"embed"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/xcutr-service/pkg/errors"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type ClickhouseClient struct {
	conn driver.Conn
}

func New(cfg *config.Config) (*ClickhouseClient, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{net.JoinHostPort(
			cfg.Secrets.Clickhouse.Host,
			cfg.Secrets.Clickhouse.Port,
		)},
		Auth: clickhouse.Auth{
			Database: cfg.Secrets.Clickhouse.Database,
			Username: cfg.Secrets.Clickhouse.Username,
			Password: cfg.Secrets.Clickhouse.Password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with clickhouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return &ClickhouseClient{
		conn: conn,
	}, nil
}

func (ch *ClickhouseClient) Up() error {
	migrations, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name() < migrations[j].Name()
	})

	for _, migration := range migrations {
		if filepath.Ext(migration.Name()) != ".sql" {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + migration.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		if err := ch.conn.Exec(context.Background(), string(content)); err != nil {
			return fmt.Errorf("failed to apply migration: %w", err)
		}
	}

	return nil
}

func (ch *ClickhouseClient) WriteSession(ctx context.Context, userID, language string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	language = strings.TrimSpace(language)
	if language == "" {
		return customerrors.ErrNilArgs
	}

	if err := ch.conn.Exec(ctx, `INSERT INTO user_services (
			user_id,
			language,
		) VALUES (?, ?)`, userID, language); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}
