package web

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
)

type ListHandlersHandler struct {
	DB *sql.DB
}

func (h *ListHandlersHandler) Route() (string, string) {
	return http.MethodGet, "/handlers"
}

func (h *ListHandlersHandler) ServeHTTP(ctx *gin.Context) error {
	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			h.key,
			h.name,
			h.handler_type,
			h.type_name,
			a.key,
			a.name,
			COUNT(DISTINCT m.type_name)
		FROM docserve.handler AS h
		INNER JOIN docserve.application AS a
		ON a.key = h.application_key
		LEFT JOIN docserve.handler_message AS m
		ON m.handler_key = h.key
		GROUP BY h.key, a.key
		ORDER BY h.name, a.name`,
	)
	if err != nil {
		return err
	}

	tc := templates.ListHandlersContext{
		Context: templates.Context{
			Title:          "Handlers",
			ActiveMenuItem: templates.HandlersMenuItem,
		},
	}

	for rows.Next() {
		var tr templates.HandlerListRow

		if err := rows.Scan(
			&tr.HandlerKey,
			&tr.HandlerName,
			&tr.HandlerType,
			&tr.HandlerTypeName,
			&tr.AppKey,
			&tr.AppName,
			&tr.MessageCount,
		); err != nil {
			return err
		}

		tc.Table = append(tc.Table, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "handlers.html", tc)

	return nil
}
