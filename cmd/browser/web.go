package main

import (
	"crypto/rsa"
	"database/sql"
	"net/http"

	"github.com/dogmatiq/browser/githubx"
	"github.com/dogmatiq/browser/web"
)

func init() {
	provide(func(
		c *githubx.Connector,
		db *sql.DB,
		pk *rsa.PrivateKey,
	) http.Handler {
		return web.NewRouter(
			version,
			c,
			pk,
			db,
		)
	})
}
