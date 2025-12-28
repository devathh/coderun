package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	ssopb "github.com/devathh/coderun/sso-service/api/sso/v1"
	"github.com/devathh/coderun/sso-service/internal/domain/auth"
	"github.com/devathh/coderun/sso-service/internal/domain/user"
	"github.com/devathh/coderun/sso-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"github.com/google/uuid"
)

type ssoService struct {
	cfg        *config.Config
	log        *slog.Logger
	userMongo  user.MongoRepository
	authCache  auth.AuthRedis
	jwtManager auth.JWTManager
}

type SSOService interface {
	Register(context.Context, *ssopb.RegisterRequest) (*ssopb.Token, error)
	Login(context.Context, *ssopb.LoginRequest) (*ssopb.Token, error)
	Refresh(context.Context, *ssopb.RefreshRequest) (*ssopb.Token, error)
	UpdateUser(context.Context, *ssopb.UpdateRequest) error
	GetUserByID(context.Context, *ssopb.GetByIDRequest) (*ssopb.User, error)
	GetSelf(context.Context) (*ssopb.User, error)
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	userMongo user.MongoRepository,
	authCache auth.AuthRedis,
	jwtManager auth.JWTManager,
) (SSOService, error) {
	if cfg == nil || log == nil {
		return nil, customerrors.ErrNilArgs
	}

	return &ssoService{
		cfg:        cfg,
		log:        log,
		userMongo:  userMongo,
		authCache:  authCache,
		jwtManager: jwtManager,
	}, nil
}

func (s *ssoService) GetUserByID(ctx context.Context, req *ssopb.GetByIDRequest) (*ssopb.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, customerrors.ErrNilRequest
	}

	userID := strings.TrimSpace(req.GetUserId())
	if userID == "" {
		return nil, customerrors.ErrInvalidRequest
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("start to get user from mongo by id", slog.String("id", userID))
	user, err := s.userMongo.GetByID(ctxTimeout, userID)
	if err != nil {
		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return nil, err
		}

		s.log.Error("failed to get user by id", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	s.log.Debug("user found successfully")
	return &ssopb.User{
		Id:       user.ID().String(),
		Email:    string(user.Email()),
		Username: user.Username(),
	}, nil
}

func (s *ssoService) Login(ctx context.Context, req *ssopb.LoginRequest) (*ssopb.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, customerrors.ErrNilRequest
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("start to search user by email")
	user, err := s.userMongo.GetByEmail(ctxTimeout, req.GetEmail())
	if err != nil {
		if errors.Is(err, customerrors.ErrNilArgs) {
			return nil, err
		}

		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return nil, err
		}

		s.log.Error("failed to get user by email", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	s.log.Debug("check user password")
	if !user.Password().Check(req.GetPassword()) {
		return nil, customerrors.ErrInvalidCredentials
	}

	s.log.Debug("start to create user's session")
	access, refresh, err := s.createSession(ctxTimeout, user.ID(), string(user.Email()))
	if err != nil {
		s.log.Error("failed to create session", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	return &ssopb.Token{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (s *ssoService) Refresh(ctx context.Context, req *ssopb.RefreshRequest) (*ssopb.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, customerrors.ErrNilRequest
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("start to search user's session by refresh token")
	session, err := s.authCache.GetSession(ctxTimeout, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, customerrors.ErrNoSessions) {
			return nil, err
		}

		s.log.Error("failed to get session", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	s.log.Debug("creating user's session")
	newAccess, newRefresh, err := s.createSession(ctxTimeout, session.UserID(), session.Email())
	if err != nil {
		s.log.Error("failed to create new session", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	go s.clearOldSession(context.Background(), req.GetRefreshToken())

	return &ssopb.Token{
		Access:  newAccess,
		Refresh: newRefresh,
	}, nil
}

func (s *ssoService) Register(ctx context.Context, req *ssopb.RegisterRequest) (*ssopb.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, customerrors.ErrNilRequest
	}

	s.log.Debug("creating domain user")
	user, err := s.createUser(req)
	if err != nil {
		return nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("start to save user")
	saved, err := s.userMongo.Save(ctxTimeout, user)
	if err != nil {
		if errors.Is(err, customerrors.ErrUserAlreadyRegistered) {
			return nil, err
		}

		s.log.Error("failed to save user into mongo", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	s.log.Debug("creating session")
	access, refresh, err := s.createSession(ctxTimeout, saved.ID(), string(saved.Email()))
	if err != nil {
		s.log.Error("failed to create session", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	return &ssopb.Token{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (s *ssoService) UpdateUser(ctx context.Context, req *ssopb.UpdateRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	userID, err := s.getUserID(ctx)
	if err != nil {
		return err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("creating domain to update")
	userUpd := user.From(userID, req.GetUsername(), user.Email(""), user.Password(""))

	s.log.Debug("starting update user")
	if err := s.userMongo.Update(ctxTimeout, userUpd); err != nil {
		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return err
		}

		s.log.Error("failed to update user", slog.String("error", err.Error()))
		return customerrors.ErrInternalServer
	}

	return nil
}

func (s *ssoService) GetSelf(ctx context.Context) (*ssopb.User, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("start to search user by id", slog.String("id", userID.String()))
	user, err := s.userMongo.GetByID(ctxTimeout, userID.String())
	if err != nil {
		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return nil, err
		}

		s.log.Error("failed to get user by id (self)", slog.String("error", err.Error()))
		return nil, customerrors.ErrInternalServer
	}

	return &ssopb.User{
		Id:       userID.String(),
		Email:    string(user.Email()),
		Username: user.Username(),
	}, nil
}

func (s *ssoService) getUserID(ctx context.Context) (uuid.UUID, error) {
	rawID := ctx.Value(auth.CtxKey("user_id"))
	if rawID == nil {
		return uuid.Nil, customerrors.ErrInvalidUserID
	}

	if id, ok := rawID.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, customerrors.ErrInvalidUserID
}

func (s *ssoService) createUser(req *ssopb.RegisterRequest) (*user.User, error) {
	email, err := user.NewEmail(req.GetEmail())
	if err != nil {
		return nil, err
	}

	password, err := user.NewPassword(req.GetPassword())
	if err != nil {
		return nil, err
	}

	return user.New(req.GetUsername(), email, password)
}

func (s *ssoService) createSession(ctx context.Context, userID uuid.UUID, email string) (string, string, error) {
	access, refresh, err := s.jwtManager.GeneratePair(userID, email)
	if err != nil {
		return "", "", err
	}

	session, err := auth.NewSession(userID, email)
	if err != nil {
		return "", "", err
	}

	if err := s.authCache.CreateSession(ctx, refresh, session); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *ssoService) clearOldSession(ctx context.Context, refresh string) {
	s.log.Debug("start to clear old session")
	if err := s.authCache.DeleteSession(ctx, refresh); err != nil {
		s.log.Error("failed to clear session", slog.String("error", err.Error()))
	}
}
