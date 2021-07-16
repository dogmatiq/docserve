package main

import (
	"database/sql"

	"github.com/dogmatiq/browser/analyzer"
	"github.com/dogmatiq/browser/githubx"
	"github.com/dogmatiq/dodeca/logging"
)

func init() {
	provide(func(
		db *sql.DB,
		c *githubx.Connector,
		l logging.Logger,
	) *analyzer.Analyzer {
		return &analyzer.Analyzer{
			DB:        db,
			Connector: c,
			Logger:    l,
		}
	})

	provide(func(
		db *sql.DB,
		l logging.Logger,
	) *analyzer.Remover {
		return &analyzer.Remover{
			DB:     db,
			Logger: l,
		}
	})

	provide(func(
		a *analyzer.Analyzer,
		r *analyzer.Remover,
	) *analyzer.Orchestrator {
		return &analyzer.Orchestrator{
			Analyzer: a,
			Remover:  r,
		}
	})
}
