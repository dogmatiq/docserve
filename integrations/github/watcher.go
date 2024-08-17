package github

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/dogmatiq/browser/integrations/github/internal/githubapi"
	"github.com/dogmatiq/browser/model"
	"github.com/dogmatiq/minibus"
	"github.com/google/go-github/v63/github"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
)

// RepositoryWatcher publishes messages about the repositories that are
// accessible to the GitHub application.
type RepositoryWatcher struct {
	Client *githubapi.AppClient
	Logger *slog.Logger
}

// Run starts the watcher.
func (w *RepositoryWatcher) Run(ctx context.Context) (err error) {
	minibus.Subscribe[*github.InstallationEvent](ctx)
	minibus.Subscribe[*github.InstallationRepositoriesEvent](ctx)
	minibus.Subscribe[*github.RepositoryEvent](ctx)
	minibus.Ready(ctx)

	if err := githubapi.Range(
		ctx,
		w.Client.REST().Apps.ListInstallations,
		w.addInstallation,
	); err != nil {
		return err
	}

	for m := range minibus.Inbox(ctx) {
		if err := w.handleMessage(ctx, m); err != nil {
			return err
		}
	}

	return ctx.Err()
}

func (w *RepositoryWatcher) handleMessage(
	ctx context.Context,
	m any,
) error {
	switch m := m.(type) {
	case *github.InstallationEvent:
		return w.handleInstallationEvent(ctx, m)
	case *github.InstallationRepositoriesEvent:
		return w.handleInstallationRepositoriesEvent(ctx, m)
	case *github.RepositoryEvent:
		return w.handleRepositoryEvent(ctx, m)
	default:
		panic("unexpected message type")
	}
}

func (w *RepositoryWatcher) handleInstallationEvent(
	ctx context.Context,
	m *github.InstallationEvent,
) error {
	switch m.GetAction() {
	case "created":
		return w.addInstallation(ctx, m.GetInstallation())
	case "deleted":
		return w.removeInstallation(ctx, m.GetInstallation(), m.Repositories)
	default:
		return nil
	}
}

func (w *RepositoryWatcher) handleInstallationRepositoriesEvent(
	ctx context.Context,
	m *github.InstallationRepositoriesEvent,
) error {
	c := w.Client.InstallationClient(m.GetInstallation().GetID())

	if err := w.addRepos(ctx, c, m.RepositoriesAdded...); err != nil {
		return err
	}

	if err := w.removeRepos(ctx, c, m.RepositoriesRemoved...); err != nil {
		return err
	}

	return nil
}

func (w *RepositoryWatcher) handleRepositoryEvent(
	ctx context.Context,
	m *github.RepositoryEvent,
) error {
	c := w.Client.InstallationClient(m.GetInstallation().GetID())

	switch m.GetAction() {
	case "unarchived":
		return w.addRepos(ctx, c, m.Repo)
	case "archived":
		return w.removeRepos(ctx, c, m.Repo)
	default:
		return nil
	}
}

func (w *RepositoryWatcher) addInstallation(
	ctx context.Context,
	in *github.Installation,
) error {
	c := w.Client.InstallationClient(in.GetID())

	w.Logger.InfoContext(
		ctx,
		"github app installation discovered",
		slog.Group(
			"installation",
			slog.Int64("id", in.GetID()),
			slog.String("account", in.GetAccount().GetLogin()),
		),
	)

	if err := githubapi.RangePages(
		ctx,
		func(ctx context.Context, opts *github.ListOptions) ([]*github.Repository, *github.Response, error) {
			repos, res, err := c.REST().Apps.ListRepos(ctx, opts)
			if err != nil {
				return nil, res, err
			}
			return repos.Repositories, res, nil
		},
		func(ctx context.Context, repos []*github.Repository) error {
			return w.addRepos(ctx, c, repos...)
		},
	); err != nil {
		return fmt.Errorf("unable to add %q installation: %w", in.GetAccount().GetLogin(), err)
	}

	return nil
}

func (w *RepositoryWatcher) removeInstallation(
	ctx context.Context,
	in *github.Installation,
	repos []*github.Repository,
) error {
	c := w.Client.InstallationClient(in.GetID())
	return w.removeRepos(ctx, c, repos...)
}

func (w *RepositoryWatcher) addRepos(
	ctx context.Context,
	c *githubapi.InstallationClient,
	repos ...*github.Repository,
) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, repo := range repos {
		if !repoIgnored(repo) {
			g.Go(func() error {
				return w.addRepo(ctx, c, repo)
			})
		}
	}

	return g.Wait()
}

func (w *RepositoryWatcher) addRepo(
	ctx context.Context,
	c *githubapi.InstallationClient,
	r *github.Repository,
) error {
	g := marshalRepo(r)

	if err := minibus.Send(
		ctx,
		repoFound{
			AppClientID:    w.Client.ID,
			InstallationID: c.InstallationID,
			GitHubRepo:     r,
			Repo:           g,
		},
	); err != nil {
		return err
	}

	tree, _, err := c.REST().Git.GetTree(
		ctx,
		r.GetOwner().GetLogin(),
		r.GetName(),
		"HEAD",
		true,
	)
	if err != nil {
		return fmt.Errorf("unable to read git tree: %w", err)
	}

	if tree.GetTruncated() {
		w.Logger.WarnContext(
			ctx,
			"truncated git tree results, some modules may go undetected",
			g.AsLogAttr(),
			slog.String("sha", tree.GetSHA()),
		)
	}

	for _, entry := range tree.Entries {
		if entry.GetType() != "blob" {
			continue
		}

		if filepath.Base(entry.GetPath()) != "go.mod" {
			continue
		}

		data, _, err := c.REST().Git.GetBlobRaw(
			ctx,
			r.GetOwner().GetLogin(),
			r.GetName(),
			entry.GetSHA(),
		)
		if err != nil {
			return fmt.Errorf("unable to read git blob: %w", err)
		}

		mod, err := modfile.ParseLax(
			entry.GetPath(),
			data,
			nil, /* version fixer*/
		)
		if err != nil {
			return fmt.Errorf("unable to parse go.mod: %w", err)
		}

		if err := minibus.Send(
			ctx,
			model.ModuleDiscovered{
				Repo: g,
				Module: model.Module{
					Path:    mod.Module.Mod.Path,
					Version: tree.GetSHA(),
				},
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func (w *RepositoryWatcher) removeRepos(
	ctx context.Context,
	c *githubapi.InstallationClient,
	repos ...*github.Repository,
) error {
	for _, r := range repos {
		if err := minibus.Send(
			ctx,
			repoLost{
				Repo:           marshalRepo(r),
				AppClientID:    w.Client.ID,
				InstallationID: c.InstallationID,
			},
		); err != nil {
			return err
		}
	}

	return nil
}
