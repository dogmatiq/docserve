package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/docserve/web/components"
	"github.com/gin-gonic/gin"
)

type listView struct {
	TotalRepoCount int
	TotalAppCount  int
	Handlers       []handlerSummary
}

type handlerSummary struct {
	Key          string
	Name         string
	Type         configkit.HandlerType
	Impl         components.Type
	AppKey       string
	AppName      string
	MessageCount int
}

type ListHandler struct {
	DB *sql.DB
}

func (h *ListHandler) Route() (string, string) {
	return http.MethodGet, "/handlers"
}

func (h *ListHandler) Template() string {
	return "handlers/list.html"
}

func (h *ListHandler) ActiveMenuItem() components.MenuItem {
	return components.HandlersMenuItem
}

func (h *ListHandler) View(ctx *gin.Context) (string, interface{}, error) {
	var view listView

	if err := h.loadStats(ctx, &view); err != nil {
		return "", nil, err
	}

	if err := h.loadHandlers(ctx, &view); err != nil {
		return "", nil, err
	}

	return "Handlers", view, nil
}

func (h *ListHandler) loadStats(
	ctx context.Context,
	view *listView,
) error {
	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			(SELECT COUNT(*) FROM docserve.repository),
			(SELECT COUNT(*) FROM docserve.application)`,
	)
	return row.Scan(
		&view.TotalRepoCount,
		&view.TotalAppCount,
	)
}

func (h *ListHandler) loadHandlers(
	ctx context.Context,
	view *listView,
) error {
	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			h.key,
			h.name,
			h.handler_type,
			t.package,
			t.name,
			h.is_pointer,
			t.url,
			t.docs,
			a.key,
			a.name,
			(
				SELECT COUNT(DISTINCT m.type_id)
				FROM docserve.handler_message AS m
				WHERE m.handler_key = h.key
			) AS message_count
		FROM docserve.handler AS h
		INNER JOIN docserve.type AS t
		ON t.id = h.type_id
		INNER JOIN docserve.application AS a
		ON a.key = h.application_key
		ORDER BY h.name, a.name`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var s handlerSummary

		if err := rows.Scan(
			&s.Key,
			&s.Name,
			&s.Type,
			&s.Impl.Package,
			&s.Impl.Name,
			&s.Impl.IsPointer,
			&s.Impl.URL,
			&s.Impl.Docs,
			&s.AppKey,
			&s.AppName,
			&s.MessageCount,
		); err != nil {
			return err
		}

		view.Handlers = append(view.Handlers, s)
	}

	return rows.Err()
}
