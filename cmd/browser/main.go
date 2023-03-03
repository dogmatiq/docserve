package main

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/dogmatiq/browser/analyzer"
	"github.com/dogmatiq/browser/githubx"
	"github.com/dogmatiq/dodeca/logging"
	"github.com/dogmatiq/ferrite"
	"github.com/google/go-github/v38/github"
	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
)

// version is the current version, set automatically by the makefiles.
var version string

func init() {
	rand.Seed(time.Now().UnixNano())

	// Setup a system-wide logger that includes debug messages only if the DEBUG
	// environment variable is set to "true".
	provide(func() logging.Logger {
		return &logging.StandardLogger{
			CaptureDebug: true,
		}
	})
}

func main() {
	ferrite.Init()

	invoke(func(
		c *githubx.Connector,
		o *analyzer.Orchestrator,
		h http.Handler,
		l logging.Logger,
	) error {
		ctx := context.Background()
		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			for {
				if err := o.Run(ctx); err != nil {
					if err == context.Canceled {
						return err
					}

					logging.LogString(l, err.Error())
				}
			}
		})

		g.Go(func() error {
			return githubx.ListInstallations(
				ctx,
				c.AppClient,
				func(ctx context.Context, i *github.Installation) error {
					ic, err := c.InstallationClient(ctx, i.GetID())
					if err != nil {
						return err
					}

					return githubx.ListRepos(
						ctx,
						ic,
						func(ctx context.Context, r *github.Repository) error {
							return o.EnqueueAnalyis(ctx, r.GetID())
						},
					)
				},
			)
		})

		g.Go(func() error {
			s := http.Server{
				Addr:    ":8080",
				Handler: h,
			}
			go func() {
				<-ctx.Done()
				s.Close()
			}()

			return s.ListenAndServe()
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
