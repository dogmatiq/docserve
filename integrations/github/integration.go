package github

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dogmatiq/browser/integrations/github/internal/githubapi"
	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/minibus"
)

var (
	// DependencyCatalog defines the dependencies of the GitHub integration.
	DependencyCatalog = &imbue.Catalog{}

	// EnvironmentRegistry defines the environment variables used by the GitHub
	// integration.
	EnvironmentRegistry = ferrite.NewRegistry("github", "GitHub Integration")
)

var (
	githubURL = ferrite.
			URL("GITHUB_URL", "the base URL of the GitHub API").
			Optional(ferrite.WithRegistry(EnvironmentRegistry))

	githubAppClientID = ferrite.
				String("GITHUB_APP_CLIENT_ID", "the client ID of the GitHub application used to read repository content").
				Required(ferrite.WithRegistry(EnvironmentRegistry))

	githubAppPrivateKey = ferrite.
				String("GITHUB_APP_PRIVATE_KEY", "the private key for the GitHub application used to read repository content").
				WithSensitiveContent().
				Required(ferrite.WithRegistry(EnvironmentRegistry))

	githubAppHookSecret = ferrite.
				String("GITHUB_APP_HOOK_SECRET", "the secret used to verify GitHub web-hook requests are genuine").
				WithSensitiveContent().
				Required(ferrite.WithRegistry(EnvironmentRegistry))
)

func init() {
	imbue.With1(
		DependencyCatalog,
		func(
			ctx imbue.Context,
			client *githubapi.AppClient,
		) (*AskpassServer, error) {
			return &AskpassServer{
				Client: client,
			}, nil
		},
	)

	imbue.With2(
		DependencyCatalog,
		func(
			ctx imbue.Context,
			client *githubapi.AppClient,
			logger *slog.Logger,
		) (*RepositoryWatcher, error) {
			return &RepositoryWatcher{
				Client: client,
				Logger: logger.With(
					slog.String("component", "github.watcher"),
				),
			}, nil
		},
	)

	imbue.With1(
		DependencyCatalog,
		func(
			ctx imbue.Context,
			logger *slog.Logger,
		) (*WebHookHandler, error) {
			return &WebHookHandler{
				Secret: githubAppHookSecret.Value(),
				Logger: logger.With(
					slog.String("component", "github.webhook"),
				),
			}, nil
		},
	)

	imbue.Decorate3(
		DependencyCatalog,
		func(
			ctx imbue.Context,
			funcs []minibus.Func,
			s *AskpassServer,
			w *RepositoryWatcher,
			h *WebHookHandler,
		) ([]minibus.Func, error) {
			return append(
				funcs,
				s.Run,
				w.Run,
				h.Run,
			), nil
		},
	)

	imbue.Decorate1(
		DependencyCatalog,
		func(
			ctx imbue.Context,
			mux *http.ServeMux,
			handler *WebHookHandler,
		) (*http.ServeMux, error) {
			mux.Handle("/github/hook", handler)
			return mux, nil
		},
	)

	imbue.With1(
		DependencyCatalog,
		func(
			ctx imbue.Context,
			logger *slog.Logger,
		) (*githubapi.AppClient, error) {
			content := []byte(githubAppPrivateKey.Value())
			block, _ := pem.Decode(content)
			if block == nil {
				return nil, errors.New("could not load GitHub private key: no valid PEM data found")
			}

			if block.Type != "RSA PRIVATE KEY" {
				return nil, fmt.Errorf("could not load GitHub private key: expected RSA PRIVATE KEY, found %s", block.Type)
			}

			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("could not load GitHub private key: %w", err)
			}

			baseURL, _ := githubURL.Value()

			c := &githubapi.AppClient{
				ID:      githubAppClientID.Value(),
				Key:     key,
				BaseURL: baseURL,
			}

			return c, nil
		},
	)
}
