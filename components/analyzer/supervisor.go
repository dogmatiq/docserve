package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dogmatiq/browser/messages/gomod"
	"github.com/dogmatiq/browser/worker"
	"github.com/dogmatiq/configkit/static"
	"golang.org/x/tools/go/packages"
)

// Supervisor oversees a pool of workers that perform static analysis of Go
// modules.
type Supervisor struct {
	Environment []string
	Cache       Cache
	Workers     int
	Logger      *slog.Logger
}

// Run starts the analyzer.
func (s *Supervisor) Run(ctx context.Context) error {
	return worker.RunPool(
		ctx,
		max(s.Workers, 1),
		s.analyze,
	)
}

func (s *Supervisor) analyze(
	ctx context.Context,
	workerID int,
	m gomod.ModuleDownloaded,
) (err error) {
	logger := s.Logger.With(
		slog.Group(
			"worker",
			slog.Int("id", workerID),
		),
		slog.Group(
			"module",
			slog.String("path", m.ModulePath),
			slog.String("version", m.ModuleVersion),
		),
	)

	start := time.Now()
	cached := false

	defer func() {
		if err == nil {
			if cached {
				logger.DebugContext(
					ctx,
					"module already analyzed",
					slog.Duration("elapsed", time.Since(start)),
				)
			} else {
				logger.InfoContext(
					ctx,
					"module analyzed",
					slog.Duration("elapsed", time.Since(start)),
				)
			}
		} else if ctx.Err() == nil {
			logger.ErrorContext(
				ctx,
				"module analysis failed",
				slog.Duration("elapsed", time.Since(start)),
				slog.Any("error", err),
			)
		}
	}()

	entry, ok, err := s.Cache.Load(ctx, m.ModulePath, m.ModuleVersion)

	if err != nil {
		logger.WarnContext(
			ctx,
			"unable to load analyzer cache entry",
			slog.Any("error", err),
		)
	}

	if ok {
		cached = true
	} else {
		pkgs, err := s.loadPackages(ctx, m.ModuleDir)
		if err != nil {
			return err
		}

		for _, p := range pkgs {
			for _, err := range p.Errors {
				logger.WarnContext(
					ctx,
					"package could not be loaded",
					slog.Group(
						"package",
						slog.String("path", p.PkgPath),
					),
					slog.String("error", err.Error()),
				)
			}
		}

		entry = CacheEntry{
			Apps: static.FromPackages(pkgs),
		}

		if err := s.Cache.Save(ctx, m.ModulePath, m.ModuleVersion, entry); err != nil {
			logger.WarnContext(
				ctx,
				"unable to save analyzer cache entry",
				slog.Any("error", err),
			)
		}
	}

	for _, app := range entry.Apps {
		logger.InfoContext(
			ctx,
			"dogma application discovered",
			slog.Group(
				"app",
				slog.String("key", app.Identity().Key),
				slog.String("name", app.Identity().Name),
				slog.String("type", app.TypeName()),
			),
		)
	}

	return nil
}

func (s *Supervisor) loadPackages(
	ctx context.Context,
	dir string,
) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedDeps,
		Dir: dir,
		Env: s.Environment,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("unable to load packages: %w", err)
	}

	return pkgs, nil
}
