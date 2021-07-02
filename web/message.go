package web

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
)

type MessageListHandler struct {
	DB *sql.DB
}

func (h *MessageListHandler) Route() (string, string) {
	return http.MethodGet, "/messages"
}

func (h *MessageListHandler) ServeHTTP(ctx *gin.Context) error {
	tc := templates.MessageListContext{
		Context: templates.Context{
			Title:          "Messages",
			ActiveMenuItem: templates.MessagesMenuItem,
		},
	}

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
	defer rows.Close()

	for rows.Next() {
		var tr templates.MessageRow

		if err := rows.Scan(
			&tr.MessageTypeName,
			&tr.Role,
			&tr.AppCount,
			&tr.HandlerCount,
		); err != nil {
			return err
		}

		tc.Messages = append(tc.Messages, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "message-list.html", tc)

	return nil
}
