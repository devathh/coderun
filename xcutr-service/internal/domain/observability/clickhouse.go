package observability

import "context"

type ClickhouseClient interface {
	Up() error
	WriteSession(ctx context.Context, userID, language string) error
}
