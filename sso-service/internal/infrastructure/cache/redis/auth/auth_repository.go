package authredis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devathh/coderun/sso-service/internal/domain/auth"
	"github.com/devathh/coderun/sso-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type AuthRedis struct {
	cfg    *config.Config
	client *redis.Client
}

func New(cfg *config.Config, client *redis.Client) (*AuthRedis, error) {
	if cfg == nil || client == nil {
		return nil, customerrors.ErrNilArgs
	}

	return &AuthRedis{
		cfg:    cfg,
		client: client,
	}, nil
}

func (ar *AuthRedis) CreateSession(ctx context.Context, refresh string, session *auth.Session) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	model := toModel(session)
	bytesModel, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := ar.generateKey(refresh)
	if err := ar.client.Set(ctx, key, bytesModel, ar.cfg.Secrets.Redis.RefreshTTL).Err(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (ar *AuthRedis) GetSession(ctx context.Context, refresh string) (*auth.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := ar.generateKey(refresh)
	bytesModel, err := ar.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, customerrors.ErrNoSessions
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var model SessionModel
	if err := json.Unmarshal(bytesModel, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	session, err := toDomain(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain: %w", err)
	}

	return session, nil
}

func (ar *AuthRedis) DeleteSession(ctx context.Context, refresh string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := ar.generateKey(refresh)
	if err := ar.client.Del(ctx, key).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return customerrors.ErrNoSessions
		}

		return fmt.Errorf("failed to delete redis: %w", err)
	}

	return nil
}

func (ar *AuthRedis) generateKey(refresh string) string {
	bytes := sha256.Sum256([]byte(refresh))
	return "rtk_" + hex.EncodeToString(bytes[:])
}
