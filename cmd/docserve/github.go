package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/dogmatiq/docserve/githubx"
	"github.com/dogmatiq/dodeca/config"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

func init() {
	provide(func(
		env config.Bucket,
		pk *rsa.PrivateKey,
	) *github.Client {
		c := github.NewClient(
			oauth2.NewClient(
				context.Background(),
				&githubx.AppTokenSource{
					AppID:      config.AsInt64(env, "GITHUB_APP_ID"),
					PrivateKey: pk,
				},
			),
		)

		if u := config.AsURLDefault(env, "GITHUB_URL", ""); u.String() != "" {
			// GitHub client requires path to have a trailing slash.
			if !strings.HasSuffix(u.Path, "/") {
				u.Path += "/"
			}

			c.BaseURL = u
		}

		return c
	})

	provide(func(env config.Bucket) (*rsa.PrivateKey, error) {
		content := config.AsBytes(env, "GITHUB_APP_PRIVATEKEY")
		block, _ := pem.Decode(content)
		if block == nil {
			return nil, errors.New("could not decode PEM key")
		}

		if block.Type != "RSA PRIVATE KEY" {
			return nil, fmt.Errorf("unexpected PEM content: %s", block.Type)
		}

		return x509.ParsePKCS1PrivateKey(block.Bytes)
	})
}
