package analyzer

import (
	"context"
	"database/sql"

	"github.com/dogmatiq/docserve/persistence"
	"github.com/dogmatiq/dodeca/logging"
	"github.com/google/go-github/v35/github"
)

type Remover struct {
	DB     *sql.DB
	Logger logging.Logger
}

func (rm *Remover) Remove(ctx context.Context, r *github.Repository) error {
	logging.Log(rm.Logger, "[%s] removing repository", r.GetFullName())

	tx, err := rm.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	if err := persistence.RemoveRepository(ctx, tx, r); err != nil {
		return err
	}

	return tx.Commit()
}
