package web

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
)

type HandlerListHandler struct {
	DB *sql.DB
}

func (h *HandlerListHandler) Route() (string, string) {
	return http.MethodGet, "/handlers"
}

func (h *HandlerListHandler) ServeHTTP(ctx *gin.Context) error {
	tc := templates.HandlerListContext{
		Context: templates.Context{
			Title:          "Handlers",
			ActiveMenuItem: templates.HandlersMenuItem,
		},
	}

	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			(SELECT COUNT(*) FROM docserve.repository),
			(SELECT COUNT(*) FROM docserve.application)`,
	)
	if err := row.Scan(
		&tc.TotalRepoCount,
		&tc.TotalAppCount,
	); err != nil {
		return err
	}

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
			&tr.MessageCount,
		); err != nil {
			return err
		}

		tc.Handlers = append(tc.Handlers, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "handler-list.html", tc)

	return nil
}

type HandlerViewHandler struct {
	DB *sql.DB
}

func (h *HandlerViewHandler) Route() (string, string) {
	return http.MethodGet, "/handlers/:key"
}

func (h *HandlerViewHandler) ServeHTTP(ctx *gin.Context) error {
	handlerKey := ctx.Param("key")

	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			h.key,
			h.name,
			h.handler_type,
			h.type_name,
			a.key,
			a.name
		FROM docserve.handler AS h
		INNER JOIN docserve.application AS a
		ON a.key = h.application_key
		WHERE h.key = $1`,
		handlerKey,
	)

	tc := templates.HandlerViewContext{
		Context: templates.Context{
			ActiveMenuItem: templates.ApplicationsMenuItem,
		},
	}

	if err := row.Scan(
		&tc.HandlerKey,
		&tc.HandlerName,
		&tc.HandlerType,
		&tc.HandlerTypeName,
		&tc.AppKey,
		&tc.AppName,
	); err != nil {
		if err == sql.ErrNoRows {
			ctx.HTML(http.StatusNotFound, "error-404.html", tc)
			return nil
		}

		return err
	}

	tc.Title = tc.AppName + " - Handler Details"

	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			m.type_name,
			m.role,
			m.produced,
			m.consumed
		FROM docserve.handler_message AS m
		WHERE m.handler_key = $1
		ORDER BY m.type_name`,
		handlerKey,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tr templates.MessageRow

		var produced, consumed bool
		if err := rows.Scan(
			&tr.MessageTypeName,
			&tr.Role,
			&produced,
			&consumed,
		); err != nil {
			return err
		}

		if tr.Role == message.TimeoutRole {
			tc.TimeoutMessages = append(tc.ProducedMessages, tr)
		} else if produced {
			tc.ProducedMessages = append(tc.ProducedMessages, tr)
		} else if consumed {
			tc.ConsumedMessages = append(tc.ConsumedMessages, tr)
		}
	}

	switch tc.HandlerType {
	case configkit.AggregateHandlerType, configkit.IntegrationHandlerType:
		tc.ConsumedMessageRole = message.CommandRole
		tc.ProducedMessageRole = message.EventRole
	case configkit.ProcessHandlerType:
		tc.ConsumedMessageRole = message.EventRole
		tc.ProducedMessageRole = message.CommandRole
	case configkit.ProjectionHandlerType:
		tc.ConsumedMessageRole = message.EventRole
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "handler-view.html", tc)

	return nil
}
