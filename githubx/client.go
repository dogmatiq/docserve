package githubx

import (
	"context"

	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

// NewClientForRepository returns a GitHub API client that has access to the
// given repository.
//
// c is the client for the GitHub application.
func NewClientForRepository(
	ctx context.Context,
	c *github.Client,
	r *github.Repository,
) (*github.Client, error) {
	install, _, err := c.Apps.FindRepositoryInstallationByID(ctx, r.GetID())
	if err != nil {
		return nil, err
	}

	return NewClientForInstallation(ctx, c, install), nil
}

// NewClientForInstallation returns a GitHub API client that has access to the
// given installation.
//
// c is the client for the GitHub application.
func NewClientForInstallation(
	ctx context.Context,
	c *github.Client,
	i *github.Installation,
) *github.Client {
	ic := github.NewClient(
		oauth2.NewClient(
			ctx,
			&InstallationTokenSource{
				AppClient:      c,
				InstallationID: i.GetID(),
			},
		),
	)

	ic.BaseURL = c.BaseURL
	ic.UploadURL = c.UploadURL
	ic.UserAgent = c.UserAgent

	return ic
}
