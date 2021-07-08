package main

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web"
	"golang.org/x/oauth2"
)

func init() {
	provide(func(
		c *oauth2.Config,
		db *sql.DB,
	) http.Handler {
		return web.NewRouter(
			c,
			&web.ApplicationListHandler{DB: db},
			&web.ApplicationViewHandler{DB: db},
			&web.HandlerListHandler{DB: db},
			&web.HandlerViewHandler{DB: db},
			&web.MessageListHandler{DB: db},
			&web.MessageViewHandler{DB: db},
		)
	})

}
