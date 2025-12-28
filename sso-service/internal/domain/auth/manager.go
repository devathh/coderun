package auth

import "github.com/google/uuid"

type JWTManager interface {
	GenerateAccess(userID uuid.UUID, email string) (string, error)
	GenerateRefresh() (string, error)
	GeneratePair(userID uuid.UUID, email string) (string, string, error)
	Validate(tokenString string) (*CoderunClaims, error)
}
