package analyzer

import (
	"context"
)

type Cloner interface {
	Clone(ctx context.Context) (string, error)
	Done()
}
