package web

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
)

type ListApplicationsHandler struct {
	DB *sql.DB
}

func (h *ListApplicationsHandler) Route() (string, string) {
	return http.MethodGet, "/"
}

func (h *ListApplicationsHandler) ServeHTTP(ctx *gin.Context) error {
	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			a.key,
			a.name,
			a.type_name,
			COUNT(DISTINCT h.key),
			COUNT(DISTINCT m.type_name)
		FROM docserve.application AS a
		LEFT JOIN docserve.handler AS h
		ON h.application_key = a.key
		LEFT JOIN docserve.handler_message AS m
		ON m.handler_key = h.key
		GROUP BY a.key
		ORDER BY a.name`,
	)
	if err != nil {
		return err
	}

	tc := templates.ListApplicationsContext{
		Context: templates.Context{
			Title:          "Applications",
			ActiveMenuItem: templates.ApplicationsMenuItem,
		},
	}

	for rows.Next() {
		var tr templates.ApplicationListRow

		if err := rows.Scan(
			&tr.AppKey,
			&tr.AppName,
			&tr.AppTypeName,
			&tr.HandlerCount,
			&tr.MessageCount,
		); err != nil {
			return err
		}

		tc.Table = append(tc.Table, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "applications.html", tc)

	return nil
}
