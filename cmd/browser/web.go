package main

import (
	"crypto/rsa"
	"database/sql"
	"net/http"

	"github.com/dogmatiq/browser/analyzer"
	"github.com/dogmatiq/browser/githubx"
	"github.com/dogmatiq/browser/web"
	"github.com/dogmatiq/dodeca/config"
)

func init() {
	provide(func(
		env config.Bucket,
		c *githubx.Connector,
		o *analyzer.Orchestrator,
		db *sql.DB,
		pk *rsa.PrivateKey,
	) http.Handler {
		return web.NewRouter(
			version,
			c,
			o,
			pk,
			config.AsBytes(env, "GITHUB_HOOK_SECRET"),
			db,
		)
	})
}
