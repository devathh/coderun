package xcutrcontainer

import (
	"time"

	customerrors "github.com/devathh/coderun/xcutr-service/pkg/errors"
	"github.com/google/uuid"
)

type Container struct {
	id          uuid.UUID
	language    Lang
	files       []File
	maxTimeout  time.Duration
	containerID string
}

func New(lang Lang, files []File, maxTimeout time.Duration) (*Container, error) {
	if len(files) < 1 {
		return nil, customerrors.ErrNoFiles
	}

	return &Container{
		id:          uuid.New(),
		language:    lang,
		files:       files,
		maxTimeout:  maxTimeout,
		containerID: "",
	}, nil
}

func From(
	id uuid.UUID,
	lang Lang,
	files []File,
	maxTimeout time.Duration,
	containerID string,
) *Container {
	return &Container{
		id:          id,
		language:    lang,
		files:       files,
		maxTimeout:  maxTimeout,
		containerID: containerID,
	}
}

func (c *Container) ID() uuid.UUID {
	return c.id
}

func (c *Container) Files() []File {
	files := make([]File, len(c.files))
	copy(files, c.files)
	return files
}

func (c *Container) Lang() Lang {
	return c.language
}

func (c *Container) MaxTimeout() time.Duration {
	return c.maxTimeout
}

func (c *Container) ContID() string {
	return c.containerID
}
