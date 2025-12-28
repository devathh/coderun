package handlers

import (
	"context"
	"errors"

	ssopb "github.com/devathh/coderun/sso-service/api/sso/v1"
	"github.com/devathh/coderun/sso-service/internal/application/services"
	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerAPI struct {
	ssopb.UnimplementedSSOServer
	service services.SSOService
}

func New(service services.SSOService) *ServerAPI {
	return &ServerAPI{
		service: service,
	}
}

func (api *ServerAPI) GetUserByID(ctx context.Context, req *ssopb.GetByIDRequest) (*ssopb.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := api.service.GetUserByID(ctx, req)
	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidRequest) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (api *ServerAPI) Login(ctx context.Context, req *ssopb.LoginRequest) (*ssopb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := api.service.Login(ctx, req)
	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, customerrors.ErrNilArgs) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (api *ServerAPI) Refresh(ctx context.Context, req *ssopb.RefreshRequest) (*ssopb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := api.service.Refresh(ctx, req)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoSessions) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (api *ServerAPI) Register(ctx context.Context, req *ssopb.RegisterRequest) (*ssopb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := api.service.Register(ctx, req)
	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidEmail) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, customerrors.ErrInvalidPassword) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, customerrors.ErrUserAlreadyRegistered) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (api *ServerAPI) UpdateUser(ctx context.Context, req *ssopb.UpdateRequest) (*ssopb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	err := api.service.UpdateUser(ctx, req)
	if err != nil {
		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (api *ServerAPI) GetSelf(ctx context.Context, _ *ssopb.Empty) (*ssopb.User, error) {
	resp, err := api.service.GetSelf(ctx)
	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidUserID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, customerrors.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}
