package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/dogmatiq/browser/messages/repo"
	"github.com/google/go-github/v63/github"
)

type repoFound struct {
	Repo repo.Repo

	AppClientID    string
	InstallationID int64
	GitHubRepo     *github.Repository
}

func (m repoFound) FoundRepo() repo.Repo {
	return m.Repo
}

func (m repoFound) LogTo(ctx context.Context, logger *slog.Logger) {
	repo.LogFound(ctx, m, logger)
}

type repoLost struct {
	Repo repo.Repo

	AppClientID    string
	InstallationID int64
	GitHubRepo     *github.Repository
}

func (m repoLost) LostRepo() repo.Repo {
	return m.Repo
}

func (m repoLost) LogTo(ctx context.Context, logger *slog.Logger) {
	repo.LogLost(ctx, m, logger)
}

// repoIgnored returns true if the repository should be ignored.
func repoIgnored(r *github.Repository) bool {
	return r.GetIsTemplate() ||
		r.GetArchived() ||
		r.GetFork()
}

// marshalRepo produces the generic representation of a GitHub repository.
func marshalRepo(r *github.Repository) repo.Repo {
	u, err := url.Parse(*r.HTMLURL)
	if err != nil {
		panic(err)
	}

	source := "github"
	if u.Host != "github.com" {
		source = fmt.Sprintf("github-enterprise-server@%s", u.Host)
	}

	return repo.Repo{
		Source: source,
		ID:     strconv.FormatInt(r.GetID(), 10),
		Name:   r.GetFullName(),
	}
}
