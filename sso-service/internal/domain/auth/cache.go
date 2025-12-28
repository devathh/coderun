package auth

import "context"

type AuthRedis interface {
	CreateSession(ctx context.Context, refresh string, session *Session) error
	GetSession(ctx context.Context, refresh string) (*Session, error)
	DeleteSession(ctx context.Context, refresh string) error
}
