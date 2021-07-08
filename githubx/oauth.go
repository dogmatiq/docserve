package githubx

import (
	"context"
	"net/url"

	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

// NewOAuthEndpoint returns the OAuth endpoint configuration to use for the
// GitHub installation at the given URL.
//
// If u is empty, the configuration for github.com is returned.
func NewOAuthEndpoint(u *url.URL) oauth2.Endpoint {
	if u.String() == "" {
		u, _ = url.Parse("https://github.com")
	} else {
		u, _ = url.Parse(u.String()) // clone
	}

	var ep oauth2.Endpoint

	u.Path = "/login/oauth/authorize"
	ep.AuthURL = u.String()

	u.Path = "/login/oauth/access_token"
	ep.TokenURL = u.String()

	return ep
}

func NewClientForUser(c *oauth2.Config, t *oauth2.Token) *github.Client {
	client := github.NewClient(
		c.Client(
			context.Background(),
			t,
		),
	)

	u, _ := url.Parse(c.Endpoint.AuthURL)
	u.Path = "/"

	if u.Host != "github.com" {
		client.BaseURL = u
	}

	return client
}
