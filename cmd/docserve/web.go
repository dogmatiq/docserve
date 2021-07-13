package main

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/githubx"
	"github.com/dogmatiq/docserve/web"
	"github.com/dogmatiq/docserve/web/pages/applications"
	"github.com/dogmatiq/docserve/web/pages/handlers"
	"github.com/dogmatiq/docserve/web/pages/messages"
)

func init() {
	provide(func(
		c *githubx.Connector,
		db *sql.DB,
	) http.Handler {
		return web.NewRouter(
			c,
			&applications.ListHandler{DB: db},
			&applications.DetailsHandler{DB: db},
			&handlers.ListHandler{DB: db},
			&handlers.DetailsHandler{DB: db},
			&messages.ListHandler{DB: db},
			&messages.DetailsHandler{DB: db},
		)
	})
}
