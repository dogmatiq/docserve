package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/dogmatiq/browser/components/internal/worker"
	"github.com/dogmatiq/browser/messages"
	"github.com/dogmatiq/configkit/static"
	"golang.org/x/tools/go/packages"
)

// Supervisor oversees a pool of workers that perform static analysis of Go
// modules.
type Supervisor struct {
	Environment []string
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
	m messages.GoModuleDownloaded,
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
	logger.DebugContext(ctx, "analyzing go module")

	defer func() {
		if err == nil {
			logger.InfoContext(
				ctx,
				"analyzed go module",
				slog.Duration("elapsed", time.Since(start)),
			)
		} else if ctx.Err() == nil {
			logger.ErrorContext(
				ctx,
				"unable to analyze go module",
				slog.Duration("elapsed", time.Since(start)),
				slog.Any("error", err),
			)
		}
	}()

	pkgs, err := s.loadPackages(ctx, m.ModuleDir)
	if err != nil {
		return err
	}

	for _, p := range pkgs {
		s.analyzePackage(ctx, p, logger)

		if ctx.Err() != nil {
			return ctx.Err()
		}
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

func (s *Supervisor) analyzePackage(
	ctx context.Context,
	p *packages.Package,
	logger *slog.Logger,
) {
	logger = logger.With(
		slog.Group(
			"package",
			slog.String("path", p.PkgPath),
		),
	)

	start := time.Now()
	logger.DebugContext(
		ctx,
		"analyzing go package",
	)
	defer func() {
		logger.InfoContext(
			ctx,
			"analyzed go package",
			slog.Duration("elapsed", time.Since(start)),
		)
	}()

	for _, err := range p.Errors {
		logger.WarnContext(
			ctx,
			"error while loading package",
			slog.String("error", err.Error()),
		)
	}

	defer func() {
		if p := recover(); p != nil {
			logger.WarnContext(
				ctx,
				"error while analyzing package",
				slog.String("error", fmt.Sprint(p)),
				slog.String("trace", string(debug.Stack())),
			)
		}
	}()

	apps := static.FromPackages([]*packages.Package{p})

	for _, app := range apps {
		logger.InfoContext(
			ctx,
			"discovered dogma application",
			slog.Group(
				"app",
				slog.String("key", app.Identity().Key),
				slog.String("name", app.Identity().Name),
			),
		)
	}
}
