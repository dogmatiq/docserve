package main

import (
	"database/sql"

	"github.com/dogmatiq/docserve/analyzer"
	"github.com/dogmatiq/dodeca/logging"
	"github.com/google/go-github/v35/github"
)

func init() {
	provide(func(
		db *sql.DB,
		c *github.Client,
		l logging.Logger,
	) *analyzer.Analyzer {
		return &analyzer.Analyzer{
			DB:           db,
			GitHubClient: c,
			Logger:       l,
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
