package xcutrcontainer

import (
	"context"

	xcutrlog "github.com/devathh/coderun/xcutr-service/internal/domain/log"
)

type ContainerRepository interface {
	Run(context.Context, *Container) (*Container, error)
	Delete(context.Context, string) error
	GetLogs(context.Context, string, chan<- *xcutrlog.Log) error
}
