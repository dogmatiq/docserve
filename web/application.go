package web

import (
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
)

type ApplicationListHandler struct {
	DB *sql.DB
}

func (h *ApplicationListHandler) Route() (string, string) {
	return http.MethodGet, "/"
}

func (h *ApplicationListHandler) ServeHTTP(ctx *gin.Context) error {
	tc := templates.ApplicationListContext{
		Context: templates.Context{
			Title:          "Applications",
			ActiveMenuItem: templates.ApplicationsMenuItem,
		},
	}

	row := h.DB.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM docserve.repository`,
	)
	if err := row.Scan(&tc.TotalRepoCount); err != nil {
		return err
	}

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
	defer rows.Close()

	for rows.Next() {
		var tr templates.ApplicationRow

		if err := rows.Scan(
			&tr.AppKey,
			&tr.AppName,
			&tr.AppTypeName,
			&tr.HandlerCount,
			&tr.MessageCount,
		); err != nil {
			return err
		}

		tc.Applications = append(tc.Applications, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "application-list.html", tc)

	return nil
}

type ApplicationViewHandler struct {
	DB *sql.DB
}

func (h *ApplicationViewHandler) Route() (string, string) {
	return http.MethodGet, "/applications/:key"
}

func (h *ApplicationViewHandler) ServeHTTP(ctx *gin.Context) error {
	appKey := ctx.Param("key")

	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			a.key,
			a.name,
			a.type_name
		FROM docserve.application AS a
		WHERE a.key = $1`,
		appKey,
	)

	tc := templates.ApplicationViewContext{
		Context: templates.Context{
			ActiveMenuItem: templates.ApplicationsMenuItem,
		},
	}

	if err := row.Scan(
		&tc.AppKey,
		&tc.AppName,
		&tc.AppTypeName,
	); err != nil {
		if err == sql.ErrNoRows {
			ctx.HTML(http.StatusNotFound, "error-404.html", tc)
			return nil
		}

		return err
	}

	tc.Title = tc.AppName + " - Application Details"

	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			h.key,
			h.name,
			h.handler_type,
			h.type_name,
			COUNT(DISTINCT m.type_name)
		FROM docserve.handler AS h
		INNER JOIN docserve.application AS a
		ON a.key = h.application_key
		LEFT JOIN docserve.handler_message AS m
		ON m.handler_key = h.key
		WHERE h.application_key = $1
		GROUP BY h.key
		ORDER BY h.name`,
		appKey,
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
			&tr.MessageCount,
		); err != nil {
			return err
		}

		tc.Handlers = append(tc.Handlers, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = h.DB.QueryContext(
		ctx,
		`SELECT
			m.type_name,
			m.role,
			COUNT(DISTINCT h.key)
		FROM docserve.handler_message AS m
		INNER JOIN docserve.handler AS h
		ON h.key = m.handler_key
		WHERE h.application_key = $1
		GROUP BY m.type_name, m.role
		ORDER BY m.type_name`,
		appKey,
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
			&tr.HandlerCount,
		); err != nil {
			return err
		}

		tc.Messages = append(tc.Messages, tr)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "application-view.html", tc)

	return nil
}
