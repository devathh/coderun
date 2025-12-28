package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	xcutrpb "github.com/devathh/coderun/xcutr-service/api/xcutr/v1"
	"github.com/devathh/coderun/xcutr-service/internal/domain/auth"
	xcutrcontainer "github.com/devathh/coderun/xcutr-service/internal/domain/container"
	xcutrlog "github.com/devathh/coderun/xcutr-service/internal/domain/log"
	"github.com/devathh/coderun/xcutr-service/internal/domain/observability"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/xcutr-service/pkg/errors"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type xcutrService struct {
	cfg      *config.Config
	log      *slog.Logger
	contRepo xcutrcontainer.ContainerRepository
	lang     map[string]xcutrcontainer.Lang
	chClient observability.ClickhouseClient
}

type XcutrService interface {
	Execute(*xcutrpb.ExecutionRequest, grpc.ServerStreamingServer[xcutrpb.Log]) error
}

func New(cfg *config.Config, log *slog.Logger, contRepo xcutrcontainer.ContainerRepository, chClient observability.ClickhouseClient) (XcutrService, error) {
	if cfg == nil || log == nil || contRepo == nil {
		return nil, customerrors.ErrNilArgs
	}

	return &xcutrService{
		cfg:      cfg,
		log:      log,
		contRepo: contRepo,
		lang: map[string]xcutrcontainer.Lang{
			"golang": xcutrcontainer.NewLang(xcutrcontainer.GO),
			"python": xcutrcontainer.NewLang(xcutrcontainer.PYTHON),
		},
		chClient: chClient,
	}, nil
}

func (x *xcutrService) Execute(req *xcutrpb.ExecutionRequest, stream grpc.ServerStreamingServer[xcutrpb.Log]) error {
	ctx := stream.Context()

	x.log.Debug("start to run the service")
	if err := x.goService(ctx, req, stream); err != nil {
		return err
	}

	if x.cfg.Features.ClickhouseEnable {
		userID, err := x.getUserID(stream.Context())
		if err != nil {
			x.log.Warn("failed to get user id from context", slog.String("error", err.Error()))
			return nil
		}

		go x.writeClckhouse(userID.String(), req.GetLanguage())
	}

	return nil
}

func (x *xcutrService) goService(ctx context.Context, req *xcutrpb.ExecutionRequest, stream grpc.ServerStreamingServer[xcutrpb.Log]) error {
	cont, err := x.createCont(req)
	if err != nil {
		return err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, cont.MaxTimeout())
	defer cancel()

	// Build n' start the container
	x.log.Debug("start to run container")
	runningCont, err := x.contRepo.Run(ctxTimeout, cont)
	if err != nil {
		x.log.Error("failed to run container", slog.String("error", err.Error()))
		return customerrors.ErrInternalServer
	}

	// After all, delete the container
	defer func() {
		x.log.Debug("delete the container", slog.String("container_id", runningCont.ContID()))
		if err := x.contRepo.Delete(ctx, runningCont.ContID()); err != nil {
			x.log.Warn("failed to delete container", slog.String("error", err.Error()))
		}
	}()

	// Create server-stream.
	// just transferring logs
	// from the container to the stream
	x.log.Debug("getting logs", slog.String("container_id", runningCont.ContID()))
	logChan := make(chan *xcutrlog.Log, x.cfg.Service.Log.BufSize)
	if err := x.contRepo.GetLogs(ctxTimeout, runningCont.ContID(), logChan); err != nil {
		if errors.Is(err, customerrors.ErrNotFoundContainer) {
			return err
		}

		x.log.Error("failed to get logs from container", slog.String("error", err.Error()))
		return customerrors.ErrInternalServer
	}

	x.streamLogs(logChan, stream)

	return nil
}

func (x *xcutrService) streamLogs(logChan <-chan *xcutrlog.Log, stream grpc.ServerStreamingServer[xcutrpb.Log]) {
	for log := range logChan {
		if err := stream.Send(&xcutrpb.Log{
			Msg: log.Msg(),
		}); err != nil {
			break
		}
	}
}

func (x *xcutrService) createCont(req *xcutrpb.ExecutionRequest) (*xcutrcontainer.Container, error) {
	// Convert request's files to domain
	files := make([]xcutrcontainer.File, 0, len(req.GetFiles()))
	var mainExists bool
	for _, file := range req.GetFiles() {
		if file.GetName() == "main" {
			mainExists = true
		}

		domainFile, err := xcutrcontainer.NewFile(
			file.GetName(),
			file.GetMime(),
			file.GetBody(),
		)
		if err != nil {
			return nil, err
		}

		files = append(files, domainFile)
	}

	// Checking main file
	if !mainExists {
		return nil, customerrors.ErrNoMain
	}

	timeout := time.Duration(req.GetMaxTimeout())
	if timeout > x.cfg.Service.MaxTimeout {
		return nil, customerrors.ErrTooLargeTimeout
	}

	if _, ok := x.lang[req.GetLanguage()]; !ok {
		return nil, customerrors.ErrInvalidLang
	}

	cont, err := xcutrcontainer.New(
		x.lang[req.GetLanguage()],
		files,
		timeout,
	)
	if err != nil {
		return nil, err
	}

	return cont, nil
}

func (x *xcutrService) getUserID(ctx context.Context) (uuid.UUID, error) {
	rawID := ctx.Value(auth.CtxKey("user_id"))
	if rawID == nil {
		return uuid.Nil, customerrors.ErrInvalidToken
	}

	if userID, ok := rawID.(uuid.UUID); ok {
		return userID, nil
	}

	return uuid.Nil, customerrors.ErrInvalidToken
}

func (x *xcutrService) writeClckhouse(userID, lang string) {
	if err := x.chClient.WriteSession(context.Background(), userID, lang); err != nil {
		x.log.Warn("failed to save session into clickhouse", slog.String("error", err.Error()))
		return
	}
}
