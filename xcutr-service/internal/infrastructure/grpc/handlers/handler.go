package handlers

import (
	"errors"

	xcutrpb "github.com/devathh/coderun/xcutr-service/api/xcutr/v1"
	services "github.com/devathh/coderun/xcutr-service/internal/application/service"
	customerrors "github.com/devathh/coderun/xcutr-service/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerAPI struct {
	xcutrpb.UnimplementedXcutrServer
	service services.XcutrService
}

func NewHandler(service services.XcutrService) *ServerAPI {
	return &ServerAPI{
		service: service,
	}
}

func (sapi *ServerAPI) Execute(req *xcutrpb.ExecutionRequest, stream grpc.ServerStreamingServer[xcutrpb.Log]) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	if err := sapi.service.Execute(req, stream); err != nil {
		if errors.Is(err, customerrors.ErrNoMain) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, customerrors.ErrTooLargeTimeout) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, customerrors.ErrNoFiles) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, customerrors.ErrInvalidFilename) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, customerrors.ErrEmptyFile) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, customerrors.ErrTooLargeFile) {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
