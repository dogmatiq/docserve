package analyzer

import (
	"context"
)

type Cloner interface {
	RepositoryID() uint64
	Clone(ctx context.Context) (string, error)
	Close() error
}
