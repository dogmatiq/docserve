package main

import (
	"crypto/rsa"
	"database/sql"
	"net/http"

	"github.com/dogmatiq/browser/githubx"
	"github.com/dogmatiq/browser/web"
	"github.com/dogmatiq/browser/web/pages/applications"
	"github.com/dogmatiq/browser/web/pages/handlers"
	"github.com/dogmatiq/browser/web/pages/messages"
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
			&applications.ListHandler{DB: db},
			&applications.DetailsHandler{DB: db},
			&applications.RelationshipHandler{DB: db},
			&handlers.ListHandler{DB: db},
			&handlers.DetailsHandler{DB: db},
			&messages.ListHandler{DB: db},
			&messages.DetailsHandler{DB: db},
		)
	})
}
