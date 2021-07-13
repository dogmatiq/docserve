package applications

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/dogmatiq/docserve/web/components"
	"github.com/gin-gonic/gin"
)

// listView is the template context for list.html.
type listView struct {
	TotalRepoCount int
	Applications   []appSummary
}

// appSummary contains a summary of information about an application, for
// display within a listView.
type appSummary struct {
	Key          string
	Name         string
	Impl         components.Type
	HandlerCount int
	MessageCount int
}

// ListHandler is an implementation of web.Handler that displays a list of
// discovered Dogma applications.
type ListHandler struct {
	DB *sql.DB
}

func (h *ListHandler) Route() (string, string) {
	return http.MethodGet, "/"
}

func (h *ListHandler) Template() string {
	return "applications/list.html"
}

func (h *ListHandler) ActiveMenuItem() components.MenuItem {
	return components.ApplicationsMenuItem
}

func (h *ListHandler) View(ctx *gin.Context) (string, interface{}, error) {
	var view listView

	if err := h.loadStats(ctx, &view); err != nil {
		return "", nil, err
	}

	if err := h.loadApplications(ctx, &view); err != nil {
		return "", nil, err
	}

	return "Applications", view, nil
}

func (h *ListHandler) loadStats(ctx context.Context, view *listView) error {
	return h.DB.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM docserve.repository`,
	).Scan(&view.TotalRepoCount)
}

func (h *ListHandler) loadApplications(ctx context.Context, view *listView) error {
	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			a.key,
			a.name,
			t.package,
			t.name,
			a.is_pointer,
			t.url,
			(
				SELECT COUNT(h.key)
				FROM docserve.handler AS h
				WHERE h.application_key = a.key
			) AS handler_count,
			(
				SELECT COUNT(DISTINCT m.type_id)
				FROM docserve.handler AS h
				INNER JOIN docserve.handler_message AS m
				ON m.handler_key = h.key
				WHERE h.application_key = a.key
			) AS message_count
		FROM docserve.application AS a
		INNER JOIN docserve.type AS t
		ON t.id = a.type_id
		ORDER BY a.name`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var s appSummary

		if err := rows.Scan(
			&s.Key,
			&s.Name,
			&s.Impl.Package,
			&s.Impl.Name,
			&s.Impl.IsPointer,
			&s.Impl.URL,
			&s.HandlerCount,
			&s.MessageCount,
		); err != nil {
			return err
		}

		view.Applications = append(view.Applications, s)
	}

	return rows.Err()
}
