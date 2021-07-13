package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/dogmatiq/docserve/githubx"
	"github.com/dogmatiq/dodeca/config"
)

func init() {
	provide(func(env config.Bucket) (*githubx.Connector, error) {
		content := config.AsBytes(env, "GITHUB_APP_PRIVATEKEY")
		block, _ := pem.Decode(content)
		if block == nil {
			return nil, errors.New("could not load private key: no valid PEM data found")
		}

		if block.Type != "RSA PRIVATE KEY" {
			return nil, fmt.Errorf("could not load private key: expected RSA PRIVATE KEY, found %s", block.Type)
		}

		pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not load private key: %w", err)
		}

		return githubx.NewConnector(
			config.AsInt64(env, "GITHUB_APP_ID"),
			pk,
			config.AsString(env, "GITHUB_CLIENT_ID"),
			config.AsString(env, "GITHUB_CLIENT_SECRET"),
			config.AsURLDefault(env, "GITHUB_URL", ""),
			nil, // use default http transport
		)
	})
}
