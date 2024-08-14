package main

import (
	"log/slog"
	"runtime"

	"github.com/dogmatiq/browser/components/downloader"
	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/minibus"
)

var (
	downloadWorkers = ferrite.
		Signed[int]("DOWNLOADER_WORKERS", "the maximum number of Go modules to download concurrently").
		WithMinimum(1).
		WithDefault(runtime.NumCPU()).
		Required()
)

func init() {
	imbue.With2(
		container,
		func(
			ctx imbue.Context,
			env imbue.ByName[environment, []string],
			logger *slog.Logger,
		) (*downloader.Supervisor, error) {
			return &downloader.Supervisor{
				Environment: env.Value(),
				Workers:     downloadWorkers.Value(),
				Logger: logger.With(
					slog.String("component", "downloader"),
				),
			}, nil
		},
	)

	imbue.Decorate1(
		container,
		func(
			ctx imbue.Context,
			funcs []minibus.Func,
			s *downloader.Supervisor,
		) ([]minibus.Func, error) {
			return append(funcs, s.Run), nil
		},
	)
}
