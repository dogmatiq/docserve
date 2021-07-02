package main

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web"
)

func init() {
	provide(func(db *sql.DB) http.Handler {
		return web.NewRouter(
			&web.ListApplicationsHandler{DB: db},
			&web.ListHandlersHandler{DB: db},
			&web.ListMessagesHandler{DB: db},
		)
	})

}
