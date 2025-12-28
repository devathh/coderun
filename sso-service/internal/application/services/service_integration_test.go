package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	ssopb "github.com/devathh/coderun/sso-service/api/sso/v1"
	jwt "github.com/devathh/coderun/sso-service/internal/infrastructure/auth"
	rediscache "github.com/devathh/coderun/sso-service/internal/infrastructure/cache/redis"
	authredis "github.com/devathh/coderun/sso-service/internal/infrastructure/cache/redis/auth"
	"github.com/devathh/coderun/sso-service/internal/infrastructure/config"
	mongodb "github.com/devathh/coderun/sso-service/internal/infrastructure/persistence/mongo"
	usermongo "github.com/devathh/coderun/sso-service/internal/infrastructure/persistence/mongo/user"
	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
)

// ! mongo should be empty
func TestRegister(t *testing.T) {
	testService := buildTestSetup()

	testCases := []struct {
		Name    string
		Input   *ssopb.RegisterRequest
		WantErr error
	}{
		{Name: "base", Input: &ssopb.RegisterRequest{
			Email:    "mail@example.com",
			Username: "example_username",
			Password: "very_strong_password",
		}},

		{Name: "invalid_email", Input: &ssopb.RegisterRequest{
			Email:    "",
			Username: "example_username",
			Password: "very_strong_password",
		}, WantErr: customerrors.ErrInvalidEmail},

		{Name: "invalid_email", Input: &ssopb.RegisterRequest{
			Email:    "invalid_mail",
			Username: "example_username",
			Password: "very_strong_password",
		}, WantErr: customerrors.ErrInvalidEmail},

		{Name: "invalid_password", Input: &ssopb.RegisterRequest{
			Email:    "mail1@example.com",
			Username: "example_username",
			Password: "",
		}, WantErr: customerrors.ErrInvalidPassword},

		{Name: "invalid_username", Input: &ssopb.RegisterRequest{
			Email:    "mail2@example.com",
			Username: "",
			Password: "very_strong_password",
		}, WantErr: customerrors.ErrInvalidUsername},

		{Name: "already_registered", Input: &ssopb.RegisterRequest{
			Email:    "mail@example.com",
			Username: "atheros",
			Password: "very_strong_password",
		}, WantErr: customerrors.ErrUserAlreadyRegistered},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := testService.Register(context.Background(), tc.Input)
			if err != nil && !errors.Is(err, tc.WantErr) {
				t.Errorf("want %v, got %v", tc.WantErr, err)
			}
		})
	}
}

// ! after register tests
func TestLogin(t *testing.T) {
	testService := buildTestSetup()

	testCases := []struct {
		Name    string
		Input   *ssopb.LoginRequest
		WantErr error
	}{
		{Name: "base", Input: &ssopb.LoginRequest{
			Email:    "mail@example.com",
			Password: "very_strong_password",
		}},

		{Name: "invalid_credentials", Input: &ssopb.LoginRequest{
			Email:    "mail@example.com",
			Password: "another_password",
		}, WantErr: customerrors.ErrInvalidCredentials},

		{Name: "invalid_mail", Input: &ssopb.LoginRequest{
			Email:    "unexist@example.com",
			Password: "another_password",
		}, WantErr: customerrors.ErrUserDoesntExist},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := testService.Login(context.Background(), tc.Input)
			if err != nil && !errors.Is(err, tc.WantErr) {
				t.Errorf("want %v, got %v", tc.WantErr, err)
			}
		})
	}
}

func buildTestSetup() SSOService {
	cfg := config.MustNew("./test/test.yml")

	db, err := mongodb.Connect(cfg)
	if err != nil {
		slog.Error("failed to connect mongo", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := mongodb.CreateIndexes(db); err != nil {
		slog.Error("failed to create indexes mongo", slog.String("error", err.Error()))
		os.Exit(1)
	}

	userMongo, err := usermongo.New(slog.Default(), db)
	if err != nil {
		slog.Error("failed to create user mongo repository", slog.String("error", err.Error()))
		os.Exit(1)
	}

	redisClient, err := rediscache.Connect(cfg)
	if err != nil {
		slog.Error("failed to create connection with redis", slog.String("error", err.Error()))
		os.Exit(1)
	}

	authCache, err := authredis.New(cfg, redisClient)
	if err != nil {
		slog.Error("failed to create auth cache", slog.String("error", err.Error()))
		os.Exit(1)
	}

	jwtMngr, err := jwt.New(cfg)
	if err != nil {
		slog.Error("failed to create jwt manager", slog.String("error", err.Error()))
		os.Exit(1)
	}

	service, err := New(cfg, slog.Default(), userMongo, authCache, jwtMngr)
	if err != nil {
		slog.Error("failed to create service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return service
}
