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
	return forEachPage(
		ctx,
		func(
			ctx context.Context,
			page int,
		) (*github.Response, error) {
			installs, res, err := c.Apps.ListInstallations(ctx, &github.ListOptions{
				Page: page,
			})
			if err != nil {
				return nil, err
			}

			for _, i := range installs {
				if err := fn(ctx, i); err != nil {
					return nil, err
				}
			}

			return res, nil
		},
	)
}

func ListRepos(
	ctx context.Context,
	c *github.Client,
	fn func(context.Context, *github.Repository) error,
) error {
	return forEachPage(
		ctx,
		func(
			ctx context.Context,
			page int,
		) (*github.Response, error) {
			repos, res, err := c.Apps.ListRepos(ctx, &github.ListOptions{
				Page: page,
			})
			if err != nil {
				return nil, err
			}

			for _, r := range repos.Repositories {
				if err := fn(ctx, r); err != nil {
					return nil, err
				}
			}

			return res, err
		},
	)
}

func forEachPage(
	ctx context.Context,
	req func(context.Context, int) (*github.Response, error),
) error {
	page := 1

	for page > 0 {
		res, err := req(ctx, page)
		if err != nil {
			return err
		}

		page = res.NextPage
	}

	return nil
}
