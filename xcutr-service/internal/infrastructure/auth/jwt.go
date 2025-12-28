package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/devathh/coderun/xcutr-service/internal/domain/auth"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/xcutr-service/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	cfg    *config.Config
	public *rsa.PublicKey
}

func New(cfg *config.Config) (*JWTManager, error) {
	if cfg == nil {
		return nil, customerrors.ErrNilArgs
	}

	public, err := loadPublic(cfg)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		cfg:    cfg,
		public: public,
	}, nil
}

func (jm *JWTManager) Validate(tokenString string) (*auth.CoderunClaims, error) {
	if tokenString == "" {
		return nil, customerrors.ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &auth.CoderunClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, customerrors.ErrInvalidToken
		}

		return jm.public, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*auth.CoderunClaims); ok {
		return claims, nil
	}

	return nil, customerrors.ErrInvalidToken
}

func loadPublic(cfg *config.Config) (*rsa.PublicKey, error) {
	bytesKey, err := os.ReadFile(cfg.Secrets.JWT.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(bytesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}
