package web

import (
	"database/sql"
	"net/http"
	"strings"

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

	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			(SELECT COUNT(*) FROM docserve.repository),
			(SELECT COUNT(*) FROM docserve.application),
			(SELECT COUNT(*) FROM docserve.handler)`,
	)
	if err := row.Scan(
		&tc.TotalRepoCount,
		&tc.TotalAppCount,
		&tc.TotalHandlerCount,
	); err != nil {
		return err
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
			&tr.MessageRole,
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

type MessageViewHandler struct {
	DB *sql.DB
}

func (h *MessageViewHandler) Route() (string, string) {
	return http.MethodGet, "/messages/*typeName"
}

func (h *MessageViewHandler) ServeHTTP(ctx *gin.Context) error {
	typeName := strings.TrimPrefix(ctx.Param("typeName"), "/")

	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			m.type_name,
			string_agg(DISTINCT m.role, ', ' ORDER BY m.role)
		FROM docserve.handler_message AS m
		WHERE m.type_name = $1
		GROUP BY m.type_name`,
		typeName,
	)

	tc := templates.MessageViewContext{
		Context: templates.Context{
			ActiveMenuItem: templates.ApplicationsMenuItem,
		},
	}

	if err := row.Scan(
		&tc.MessageTypeName,
		&tc.MessageRole,
	); err != nil {
		if err == sql.ErrNoRows {
			ctx.HTML(http.StatusNotFound, "error-404.html", tc)
			return nil
		}

		return err
	}

	tc.Title = tc.MessageTypeName.Name() + " - Message Details"

	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			a.key,
			a.name,
			a.type_name
		FROM docserve.application AS a
		INNER JOIN docserve.handler AS h
		ON h.application_key = a.key
		INNER JOIN docserve.handler_message AS m
		ON m.handler_key = h.key
		WHERE m.type_name = $1
		GROUP BY a.key
		ORDER BY a.name`,
		typeName,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tr templates.ApplicationRow

		if err := rows.Scan(
			&tr.AppKey,
			&tr.AppName,
			&tr.AppTypeName,
		); err != nil {
			return err
		}

		tc.Applications = append(tc.Applications, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = h.DB.QueryContext(
		ctx,
		`SELECT
			h.key,
			h.name,
			h.handler_type,
			h.type_name,
			a.key,
			a.name,
			m.produced,
			m.consumed
		FROM docserve.handler AS h
		INNER JOIN docserve.application AS a
		ON a.key = h.application_key
		INNER JOIN docserve.handler_message AS m
		ON m.handler_key = h.key
		WHERE m.type_name = $1
		ORDER BY m.produced DESC, h.name, a.name`,
		typeName,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tr templates.HandlerRow

		if err := rows.Scan(
			&tr.HandlerKey,
			&tr.HandlerName,
			&tr.HandlerType,
			&tr.HandlerTypeName,
			&tr.AppKey,
			&tr.AppName,
			&tr.IsProducer,
			&tr.IsConsumer,
		); err != nil {
			return err
		}

		tc.Handlers = append(tc.Handlers, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "message-view.html", tc)

	return nil
}
