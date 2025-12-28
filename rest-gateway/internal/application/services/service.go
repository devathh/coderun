package services

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	ssopb "github.com/devathh/coderun/rest-gateway/api/sso/v1"
	"github.com/devathh/coderun/rest-gateway/internal/application/dto"
	"github.com/devathh/coderun/rest-gateway/internal/infrastructure/config"
	ssoclient "github.com/devathh/coderun/rest-gateway/internal/infrastructure/grpc/sso-client"
	customerrors "github.com/devathh/coderun/rest-gateway/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type restGatewayService struct {
	cfg       *config.Config
	log       *slog.Logger
	ssoClient ssoclient.SSOClient
}

type RestGatewayService interface {
	Register(context.Context, *dto.RegisterRequest) (*dto.Token, int, error)
	Login(context.Context, *dto.LoginRequest) (*dto.Token, int, error)
	Refresh(context.Context, *dto.RefreshRequest) (*dto.Token, int, error)
	UpdateUser(context.Context, *dto.UpdateRequest, string) (int, error)
	GetUserByID(context.Context, *dto.GetByIDRequest) (*dto.User, int, error)
	GetSelf(context.Context, string) (*dto.User, int, error)
}

func New(cfg *config.Config, log *slog.Logger, ssoClient ssoclient.SSOClient) RestGatewayService {
	return &restGatewayService{
		cfg:       cfg,
		log:       log,
		ssoClient: ssoClient,
	}
}

func (rgs *restGatewayService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.Token, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusGatewayTimeout, err
	}

	token, err := rgs.ssoClient.Register(ctx, &ssopb.RegisterRequest{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		errStatus, ok := status.FromError(err)
		if !ok {
			rgs.log.Error("failed to get status of error", slog.String("error", err.Error()))
			return nil, http.StatusBadGateway, customerrors.ErrInternalServer
		}

		if errStatus.Code() == codes.InvalidArgument {
			return nil, http.StatusBadRequest, errors.New(errStatus.Message())
		}

		if errStatus.Code() == codes.AlreadyExists {
			return nil, http.StatusConflict, errors.New(errStatus.Message())
		}

		rgs.log.Error("failed to do register request", slog.String("error", err.Error()))
		return nil, http.StatusBadGateway, customerrors.ErrInternalServer
	}

	return &dto.Token{
		Access:  token.Access,
		Refresh: token.Refresh,
	}, http.StatusCreated, nil
}

func (rgs *restGatewayService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.Token, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusGatewayTimeout, err
	}

	token, err := rgs.ssoClient.Login(ctx, &ssopb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		errStatus, ok := status.FromError(err)
		if !ok {
			rgs.log.Error("failed to get status of error", slog.String("error", err.Error()))
			return nil, http.StatusBadGateway, customerrors.ErrInternalServer
		}

		if errStatus.Code() == codes.Unauthenticated {
			return nil, http.StatusUnauthorized, errors.New(errStatus.Message())
		}

		if errStatus.Code() == codes.NotFound {
			return nil, http.StatusNotFound, errors.New(errStatus.Message())
		}

		if errStatus.Code() == codes.InvalidArgument {
			return nil, http.StatusBadRequest, errors.New(errStatus.Message())
		}

		rgs.log.Error("failed to do login request", slog.String("error", err.Error()))
		return nil, http.StatusBadGateway, customerrors.ErrInternalServer
	}

	return &dto.Token{
		Access:  token.Access,
		Refresh: token.Refresh,
	}, http.StatusOK, nil
}

func (rgs *restGatewayService) Refresh(ctx context.Context, req *dto.RefreshRequest) (*dto.Token, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusGatewayTimeout, err
	}

	token, err := rgs.ssoClient.Refresh(ctx, &ssopb.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		errStatus, ok := status.FromError(err)
		if !ok {
			rgs.log.Error("failed to get status of error", slog.String("error", err.Error()))
			return nil, http.StatusBadGateway, customerrors.ErrInternalServer
		}

		if errStatus.Code() == codes.Unauthenticated {
			return nil, http.StatusUnauthorized, errors.New(errStatus.Message())
		}

		rgs.log.Error("failed to do refresh request", slog.String("error", err.Error()))
		return nil, http.StatusBadGateway, customerrors.ErrInternalServer
	}

	return &dto.Token{
		Access:  token.Access,
		Refresh: token.Refresh,
	}, http.StatusOK, nil
}

func (rgs *restGatewayService) UpdateUser(ctx context.Context, req *dto.UpdateRequest, session string) (int, error) {
	if err := ctx.Err(); err != nil {
		return http.StatusGatewayTimeout, err
	}

	if err := rgs.ssoClient.UpdateUser(ctx, &ssopb.UpdateRequest{
		Username: req.Username,
	}, session); err != nil {
		errStatus, ok := status.FromError(err)
		if !ok {
			rgs.log.Error("failed to get status of error", slog.String("error", err.Error()))
			return http.StatusBadGateway, customerrors.ErrInternalServer
		}

		if errStatus.Code() == codes.NotFound {
			return http.StatusNotFound, errors.New(errStatus.Message())
		}

		rgs.log.Error("failed to do update request", slog.String("error", err.Error()))
		return http.StatusBadGateway, customerrors.ErrInternalServer
	}

	return http.StatusOK, nil
}

func (rgs *restGatewayService) GetUserByID(ctx context.Context, req *dto.GetByIDRequest) (*dto.User, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusGatewayTimeout, err
	}

	user, err := rgs.ssoClient.GetUserByID(ctx, &ssopb.GetByIDRequest{
		UserId: req.UserID,
	})
	if err != nil {
		errStatus, ok := status.FromError(err)
		if !ok {
			rgs.log.Error("failed to get status of error", slog.String("error", err.Error()))
			return nil, http.StatusBadGateway, customerrors.ErrInternalServer
		}

		if errStatus.Code() == codes.NotFound {
			return nil, http.StatusNotFound, customerrors.ErrUserNotFound
		}

		rgs.log.Error("failed to do get user request", slog.String("error", err.Error()))
		return nil, http.StatusBadGateway, errors.New(errStatus.Message())
	}

	return &dto.User{
		ID:       user.Id,
		Email:    user.Email,
		Username: user.Username,
	}, http.StatusOK, nil
}

func (rgs *restGatewayService) GetSelf(ctx context.Context, session string) (*dto.User, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusGatewayTimeout, err
	}

	user, err := rgs.ssoClient.GetSelf(ctx, session)
	if err != nil {
		errStatus, ok := status.FromError(err)
		if !ok {
			rgs.log.Error("failed to get status of error", slog.String("error", err.Error()))
			return nil, http.StatusBadGateway, customerrors.ErrInternalServer
		}

		if errStatus.Code() == codes.NotFound {
			return nil, http.StatusNotFound, customerrors.ErrUserNotFound
		}

		rgs.log.Error("failed to do get self user request", slog.String("error", err.Error()))
		return nil, http.StatusBadGateway, errors.New(errStatus.Message())
	}

	return &dto.User{
		ID:       user.Id,
		Email:    user.Email,
		Username: user.Username,
	}, http.StatusOK, nil
}
