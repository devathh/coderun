package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CoderunClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
	Email  string
}

type CtxKey string
