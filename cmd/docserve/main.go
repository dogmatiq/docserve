package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/dogmatiq/docserve/analyzer"
	"github.com/dogmatiq/docserve/githubx"
	"github.com/dogmatiq/dodeca/config"
	"github.com/dogmatiq/dodeca/logging"
	"github.com/dogmatiq/linger"
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

		if os.Getenv("SCAN") != "" {
			g.Go(func() error {
				for {
					if err := githubx.ListInstallations(
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
									return o.EnqueueAnalyis(ctx, r)
								},
							)
						},
					); err != nil {
						return err
					}

					if err := linger.Sleep(ctx, 10*time.Minute); err != nil {
						return err
					}
				}
			})
		}

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
