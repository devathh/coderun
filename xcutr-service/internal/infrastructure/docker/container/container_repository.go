package containerdocker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/containerd/errdefs"
	xcutrcontainer "github.com/devathh/coderun/xcutr-service/internal/domain/container"
	xcutrlog "github.com/devathh/coderun/xcutr-service/internal/domain/log"
	"github.com/devathh/coderun/xcutr-service/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/xcutr-service/pkg/errors"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type option struct {
	image string
	cmd   []string
}

type ContainerRepository struct {
	cfg     *config.Config
	cli     *client.Client
	options map[int]option
}

func New(cfg *config.Config, cli *client.Client) (*ContainerRepository, error) {
	if cli == nil || cfg == nil {
		return nil, customerrors.ErrNilArgs
	}
	return &ContainerRepository{
		cfg: cfg,
		cli: cli,
		options: map[int]option{
			int(xcutrcontainer.GO): {
				image: cfg.Secrets.Docker.ImageGo,
				cmd:   []string{"go", "run", "./main.go"},
			},
			int(xcutrcontainer.PYTHON): {
				image: cfg.Secrets.Docker.ImagePython,
				cmd:   []string{"python", "./main.py"},
			},
		},
	}, nil
}

func (cr *ContainerRepository) Run(ctx context.Context, domainContainer *xcutrcontainer.Container) (*xcutrcontainer.Container, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := cr.pullImage(ctx, cr.cfg.Secrets.Docker.ImageGo); err != nil {
		return nil, err
	}

	containerName := fmt.Sprintf("%s-%s", domainContainer.ID().String(), domainContainer.Lang().String())

	resp, err := cr.cli.ContainerCreate(ctx, &container.Config{
		WorkingDir: "/",
		Image:      cr.options[domainContainer.Lang().Value()].image,
		Cmd:        cr.options[domainContainer.Lang().Value()].cmd,
		Tty:        false,
	}, nil, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := cr.copyFiles(ctx, domainContainer.Files(), resp.ID, "/"); err != nil {
		return nil, err
	}

	if err := cr.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return xcutrcontainer.From(
		domainContainer.ID(),
		domainContainer.Lang(),
		domainContainer.Files(),
		domainContainer.MaxTimeout(),
		resp.ID,
	), nil
}

func (cr *ContainerRepository) Delete(ctx context.Context, containerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := cr.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return customerrors.ErrNotFoundContainer
		}

		return fmt.Errorf("failed to inspect container: %v", err)
	}

	if err := cr.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
}

func (cr *ContainerRepository) GetLogs(ctx context.Context, containerID string, logChan chan<- *xcutrlog.Log) error {
	reader, err := cr.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "all",
	})
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}

	go func() {
		defer func() {
			_ = reader.Close()
			close(logChan)
		}()

		bufReader := bufio.NewReader(reader)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			header := make([]byte, 8)
			_, err := io.ReadFull(bufReader, header)
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			size := binary.BigEndian.Uint32(header[4:8])
			if size == 0 {
				continue
			}

			data := make([]byte, size)
			_, err = io.ReadFull(bufReader, data)
			if err != nil {

				return
			}

			data = bytes.TrimRight(data, "\n")

			if len(data) == 0 {
				continue
			}

			lines := bytes.SplitSeq(data, []byte{'\n'})
			for l := range lines {
				if len(l) == 0 {
					continue
				}
				select {
				case logChan <- xcutrlog.NewLog(string(l)):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}

func (cr *ContainerRepository) pullImage(ctx context.Context, imageStr string) error {
	reader, err := cr.cli.ImagePull(ctx, imageStr, image.PullOptions{
		All: false,
	})
	if err != nil {
		return fmt.Errorf("failed to pull go image: %w", err)
	}
	defer reader.Close()

	_, _ = io.Copy(io.Discard, reader)

	return nil
}

func (cr *ContainerRepository) copyFiles(ctx context.Context, files []xcutrcontainer.File, containerID, path string) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, file := range files {
		containerPath := filepath.Join(path, fmt.Sprintf("%s.%s", file.Name(), file.Mime()))
		containerPath = filepath.ToSlash(containerPath)

		header := &tar.Header{
			Name:    containerPath,
			Mode:    0644,
			Size:    int64(len(file.Bytes())),
			ModTime: time.Now(),
			Format:  tar.FormatPAX,
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		if _, err := tw.Write(file.Bytes()); err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	err := cr.cli.CopyToContainer(ctx, containerID, "/", buf, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
		CopyUIDGID:                false,
	})
	if err != nil {
		return fmt.Errorf("failed to copy files into container: %w", err)
	}

	return nil
}
