package interceptors

import (
	"context"
	"log/slog"

	"github.com/devathh/coderun/sso-service/internal/domain/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type PackInterceptors struct {
	log         *slog.Logger
	jwtManager  auth.JWTManager
	authRequire map[string]bool
}

func New(log *slog.Logger, jwtManager auth.JWTManager, authRequire map[string]bool) *PackInterceptors {
	return &PackInterceptors{
		log:         log,
		jwtManager:  jwtManager,
		authRequire: authRequire,
	}
}

func (p *PackInterceptors) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !p.authRequire[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "failed to get metadata")
		}

		access := md.Get("session")
		if len(access) < 1 {
			return nil, status.Error(codes.Unauthenticated, "token is empty")
		}

		claims, err := p.jwtManager.Validate(access[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, auth.CtxKey("user_id"), claims.UserID)
		ctx = context.WithValue(ctx, auth.CtxKey("email"), claims.Email)

		return handler(ctx, req)
	}

}
