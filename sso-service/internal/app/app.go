package app

import (
	"fmt"
	"log/slog"
	"os"

	ssopb "github.com/devathh/coderun/sso-service/api/sso/v1"
	"github.com/devathh/coderun/sso-service/internal/application/services"
	jwt "github.com/devathh/coderun/sso-service/internal/infrastructure/auth"
	rediscache "github.com/devathh/coderun/sso-service/internal/infrastructure/cache/redis"
	authredis "github.com/devathh/coderun/sso-service/internal/infrastructure/cache/redis/auth"
	"github.com/devathh/coderun/sso-service/internal/infrastructure/config"
	server "github.com/devathh/coderun/sso-service/internal/infrastructure/grpc"
	"github.com/devathh/coderun/sso-service/internal/infrastructure/grpc/handlers"
	"github.com/devathh/coderun/sso-service/internal/infrastructure/grpc/interceptors"
	mongodb "github.com/devathh/coderun/sso-service/internal/infrastructure/persistence/mongo"
	usermongo "github.com/devathh/coderun/sso-service/internal/infrastructure/persistence/mongo/user"
	"github.com/devathh/coderun/sso-service/pkg/log"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/grpc"
)

type App struct {
	log *slog.Logger
	srv *server.Server
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

	db, err := connectMongo(cfg)
	if err != nil {
		return nil, nil, err
	}

	userMongo, err := usermongo.New(log, db)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user mongo repository: %w", err)
	}

	redisClient, err := rediscache.Connect(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	authCache, err := authredis.New(cfg, redisClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load auth redis repository: %w", err)
	}

	jwtMngr, err := jwt.New(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load jwt manager: %w", err)
	}

	service, err := services.New(cfg, log, userMongo, authCache, jwtMngr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load service: %w", err)
	}
	api := handlers.New(service)

	packInterceptors := interceptors.New(log, jwtMngr, map[string]bool{
		ssopb.SSO_UpdateUser_FullMethodName: true,
		ssopb.SSO_GetSelf_FullMethodName:    true,
	})
	pack := []grpc.UnaryServerInterceptor{
		packInterceptors.AuthInterceptor(),
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		pack...,
	))
	ssopb.RegisterSSOServer(grpcServer, api)

	server := server.New(cfg, grpcServer)

	cleanup := func() {
		if err := mongodb.Close(db); err != nil {
			log.Error("failed to close mongo connection", slog.String("error", err.Error()))
		}

		if err := rediscache.Close(redisClient); err != nil {
			log.Error("failed to close redis connection", slog.String("error", err.Error()))
		}

		log.Info("all the connections were closed")
	}

	return &App{
		log: log,
		srv: server,
	}, cleanup, nil
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.srv.Start()
}

func (a *App) Shutdown() {
	a.log.Info("server shutdown")
	a.srv.Shutdown()
}

func connectMongo(cfg *config.Config) (*mongo.Client, error) {
	db, err := mongodb.Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}
	if err := mongodb.CreateIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return db, nil
}
