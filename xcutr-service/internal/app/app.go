package app

import (
	"fmt"
	"log/slog"
	"os"

	xcutrpb "github.com/devathh/coderun/xcutr-service/api/xcutr/v1"
	services "github.com/devathh/coderun/xcutr-service/internal/application/service"
	"github.com/devathh/coderun/xcutr-service/internal/domain/observability"
	jwt "github.com/devathh/coderun/xcutr-service/internal/infrastructure/auth"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/config"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/docker"
	containerdocker "github.com/devathh/coderun/xcutr-service/internal/infrastructure/docker/container"
	grpcserver "github.com/devathh/coderun/xcutr-service/internal/infrastructure/grpc"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/grpc/handlers"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/grpc/interceptors"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/observability/clickhouse"
	"github.com/devathh/coderun/xcutr-service/pkg/log"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type App struct {
	log    *slog.Logger
	server *grpcserver.Server
}

func New() (*App, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg, err := config.New(os.Getenv("APP_CONFIG_PATH"))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.App.Env)
	if err != nil {
		return nil, fmt.Errorf("failed to setup log handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config is loaded", slog.Any("app", cfg.App), slog.Any("server", cfg.Server))

	dockerClient, err := docker.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with docker client: %w", err)
	}

	contRepo, err := containerdocker.New(cfg, dockerClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create container repository: %w", err)
	}

	var chClient observability.ClickhouseClient
	if cfg.Features.ClickhouseEnable {
		chClient, err = clickhouse.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create clickhouse client: %w", err)
		}

		if err := chClient.Up(); err != nil {
			return nil, fmt.Errorf("failed to create clickhouse client: %w", err)
		}
	}

	service, err := services.New(cfg, log, contRepo, chClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	api := handlers.NewHandler(service)

	jwtManager, err := jwt.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt manager: %w", err)
	}
	pack := interceptors.New(log, jwtManager, map[string]bool{
		xcutrpb.Xcutr_Execute_FullMethodName: true,
	})

	grpcServer := grpc.NewServer(grpc.StreamInterceptor(pack.AuthInterceptor()))
	xcutrpb.RegisterXcutrServer(grpcServer, api)

	server := grpcserver.New(cfg, grpcServer)

	return &App{
		log:    log,
		server: server,
	}, nil
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) Shutdown() {
	a.log.Info("server shutdown")
	a.server.GracefulShutdown()
}
