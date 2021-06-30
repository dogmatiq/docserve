package githubx

import (
	"context"

	"github.com/google/go-github/v35/github"
)

func ListInstallations(
	ctx context.Context,
	c *github.Client,
	fn func(context.Context, *github.Installation) error,
) error {
	options := &github.ListOptions{}

	for {
		options.Page++

		installs, _, err := c.Apps.ListInstallations(ctx, options)
		if len(installs) == 0 || err != nil {
			return err
		}

		for _, i := range installs {
			if err := fn(ctx, i); err != nil {
				return err
			}
		}
	}
}

func ListRepos(
	ctx context.Context,
	c *github.Client,
	fn func(context.Context, *github.Repository) error,
) error {
	options := &github.ListOptions{}

	for {
		options.Page++

		repos, _, err := c.Apps.ListRepos(ctx, options)
		if len(repos.Repositories) == 0 || err != nil {
			return err
		}

		for _, r := range repos.Repositories {
			if err := fn(ctx, r); err != nil {
				return err
			}
		}
	}
}
