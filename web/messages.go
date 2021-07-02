package web

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
)

type ListMessagesHandler struct {
	DB *sql.DB
}

func (h *ListMessagesHandler) Route() (string, string) {
	return http.MethodGet, "/messages"
}

func (h *ListMessagesHandler) ServeHTTP(ctx *gin.Context) error {
	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			m.type_name,
			string_agg(DISTINCT m.role, ', ' ORDER BY m.role),
			COUNT(DISTINCT a.key),
			COUNT(DISTINCT h.key)
		FROM docserve.handler_message AS m
		INNER JOIN docserve.handler AS h
		ON h.key = m.handler_key
		INNER JOIN docserve.application AS a
		ON a.key = h.application_key
		GROUP BY m.type_name
		ORDER BY m.type_name`,
	)
	if err != nil {
		return err
	}

	tc := templates.ListMessagesContext{
		Context: templates.Context{
			Title:          "Messages",
			ActiveMenuItem: templates.MessagesMenuItem,
		},
	}

	for rows.Next() {
		var tr templates.MessageListRow

		if err := rows.Scan(
			&tr.MessageTypeName,
			&tr.Role,
			&tr.AppCount,
			&tr.HandlerCount,
		); err != nil {
			return err
		}

		tc.Table = append(tc.Table, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "messages.html", tc)

	return nil
}
