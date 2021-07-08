package main

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/dogmatiq/docserve/analyzer"
	"github.com/dogmatiq/docserve/githubx"
	"github.com/dogmatiq/dodeca/config"
	"github.com/dogmatiq/dodeca/logging"
	"github.com/google/go-github/v35/github"
	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// Setup a config bucket for reading environment variables.
	provide(func() config.Bucket {
		return config.Environment()
	})

	// Setup a system-wide logger that includes debug messages only if the DEBUG
	// environment variable is set to "true".
	provide(func(env config.Bucket) logging.Logger {
		return &logging.StandardLogger{
			CaptureDebug: config.AsBoolF(env, "DEBUG"),
		}
	})
}

func main() {
	invoke(func(
		c *github.Client,
		o *analyzer.Orchestrator,
		h http.Handler,
	) error {
		ctx := context.Background()
		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			return o.Run(ctx)
		})

		g.Go(func() error {
			return githubx.ListInstallations(
				ctx,
				c,
				func(ctx context.Context, i *github.Installation) error {
					ic := githubx.NewClientForInstallation(ctx, c, i)

					return githubx.ListRepos(
						ctx,
						ic,
						func(ctx context.Context, r *github.Repository) error {
							return o.EnqueueAnalyis(ctx, r)
						},
					)
				},
			)
		})

		g.Go(func() error {
			return http.ListenAndServe(":8080", h)
		})

		return g.Wait()
	})
}

var container = dig.New()

func provide(fn interface{}) {
	if err := container.Provide(fn); err != nil {
		panic(err)
	}
}

func invoke(fn interface{}) {
	if err := container.Invoke(fn); err != nil {
		panic(err)
	}
}
