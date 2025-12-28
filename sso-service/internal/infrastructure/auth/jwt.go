package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/devathh/coderun/sso-service/internal/domain/auth"
	"github.com/devathh/coderun/sso-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	cfg     *config.Config
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func New(cfg *config.Config) (*JWTManager, error) {
	if cfg == nil {
		return nil, customerrors.ErrNilArgs
	}

	private, err := loadPrivate(cfg)
	if err != nil {
		return nil, err
	}

	public, err := loadPublic(cfg)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		cfg:     cfg,
		private: private,
		public:  public,
	}, nil
}

func (jm *JWTManager) GenerateAccess(userID uuid.UUID, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, auth.CoderunClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "shost-sso",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.cfg.Secrets.JWT.TTL)),
			Subject:   "shost-user",
		},
	})

	tokenString, err := token.SignedString(jm.private)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (jm *JWTManager) GenerateRefresh() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return "rt_" + hex.EncodeToString(buf), nil
}

// access, refresh
func (jm *JWTManager) GeneratePair(userID uuid.UUID, email string) (string, string, error) {
	access, err := jm.GenerateAccess(userID, email)
	if err != nil {
		return "", "", err
	}

	refresh, err := jm.GenerateRefresh()
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
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

func loadPrivate(cfg *config.Config) (*rsa.PrivateKey, error) {
	bytesKey, err := os.ReadFile(cfg.Secrets.JWT.PrivatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(bytesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

func loadPublic(cfg *config.Config) (*rsa.PublicKey, error) {
	bytesKey, err := os.ReadFile(cfg.Secrets.JWT.PublicPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(bytesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}
