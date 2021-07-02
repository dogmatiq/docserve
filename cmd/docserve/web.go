package main

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web"
)

func init() {
	provide(func(db *sql.DB) http.Handler {
		return web.NewRouter(
			&web.ApplicationListHandler{DB: db},
			&web.ApplicationViewHandler{DB: db},
			&web.HandlerListHandler{DB: db},
			&web.HandlerViewHandler{DB: db},
			&web.MessageListHandler{DB: db},
			&web.MessageViewHandler{DB: db},
		)
	})

}
