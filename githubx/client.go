package githubx

import (
	"context"

	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

func NewClientForRepository(
	ctx context.Context,
	c *github.Client,
	r *github.Repository,
) (*github.Client, error) {
	install, _, err := c.Apps.FindRepositoryInstallationByID(ctx, r.GetID())
	if err != nil {
		return nil, err
	}

	return github.NewClient(
		oauth2.NewClient(
			ctx,
			&InstallationTokenSource{
				AppClient:      c,
				InstallationID: install.GetID(),
			},
		),
	), nil
}
