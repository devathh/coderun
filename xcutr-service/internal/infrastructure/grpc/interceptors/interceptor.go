package interceptors

import (
	"context"
	"log/slog"

	"github.com/devathh/coderun/xcutr-service/internal/domain/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type WrapperStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ws *WrapperStream) Context() context.Context {
	return ws.ctx
}

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

func (p *PackInterceptors) AuthInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !p.authRequire[info.FullMethod] {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "failed to get metadata")
		}

		access := md.Get("session")
		if len(access) < 1 {
			return status.Error(codes.Unauthenticated, "token is empty")
		}

		claims, err := p.jwtManager.Validate(access[0])
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx := context.WithValue(ss.Context(), auth.CtxKey("user_id"), claims.UserID)

		wrappedStream := &WrapperStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}

}
