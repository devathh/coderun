package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/devathh/coderun/rest-gateway/internal/application/services"
	"github.com/devathh/coderun/rest-gateway/internal/infrastructure/config"
	ssoclient "github.com/devathh/coderun/rest-gateway/internal/infrastructure/grpc/sso-client"
	httpserver "github.com/devathh/coderun/rest-gateway/internal/infrastructure/http"
	"github.com/devathh/coderun/rest-gateway/internal/infrastructure/http/handlers"
	"github.com/devathh/coderun/rest-gateway/pkg/log"
	"github.com/joho/godotenv"
)

type App struct {
	log    *slog.Logger
	server *httpserver.Server
}

func New() (*App, func(), error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg, err := config.New(os.Getenv("APP_CONFIG_PATH"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.App.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup log handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	client, conn, err := ssoclient.Connect(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to sso: %w", err)
	}

	ssoClient, err := ssoclient.New(client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sso-client: %w", err)
	}

	service := services.New(cfg, log, *ssoClient)
	handler, err := handlers.New(cfg, service)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create handler: %w", err)
	}

	server := httpserver.New(cfg, handler)

	return &App{
			log:    log,
			server: server,
		}, func() {
			if err := ssoclient.Close(conn); err != nil {
				log.Error("failed to close sso client conn", slog.String("error", err.Error()))
			}
		}, nil
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) Shutdown() error {
	a.log.Info("server shutdown")
	return a.server.Shutdown(context.Background())
}
