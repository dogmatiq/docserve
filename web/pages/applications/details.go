package applications

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/docserve/web/components"
	"github.com/gin-gonic/gin"
)

// detailsView is the template context for details.html.
type detailsView struct {
	Key  string
	Name string
	Impl components.Type

	Handlers []handlerSummary
	Messages []messageSummary
}

// handlerSummary contains a summary of information about a handler within an
// application, for display within a detailsView.
type handlerSummary struct {
	Key          string
	Name         string
	Type         configkit.HandlerType
	Impl         components.Type
	MessageCount string
}

// handlerSummary contains a summary of information about a message used by an
// application, for display within a detailsView.
type messageSummary struct {
	Impl          components.Type
	Role          message.Role
	ProducerCount int
	ConsumerCount int
}

// DetailsHandler is an implementation of web.Handler that displays detailed
// information about a single Dogma application.
type DetailsHandler struct {
	DB *sql.DB
}

func (h *DetailsHandler) Route() (string, string) {
	return http.MethodGet, "/applications/:key"
}

func (h *DetailsHandler) Template() string {
	return "applications/details.html"
}

func (h *DetailsHandler) ActiveMenuItem() components.MenuItem {
	return components.ApplicationsMenuItem
}

func (h *DetailsHandler) View(ctx *gin.Context) (string, interface{}, error) {
	var view detailsView

	appKey := ctx.Param("key")

	if err := h.loadDetails(ctx, &view, appKey); err != nil {
		if err == sql.ErrNoRows {
			ctx.AbortWithStatus(http.StatusNotFound)
			return "", nil, nil
		}

		return "", nil, err
	}

	if err := h.loadHandlers(ctx, &view, appKey); err != nil {
		return "", nil, err
	}

	if err := h.loadMessages(ctx, &view, appKey); err != nil {
		return "", nil, err
	}

	return view.Name + " - Application Details", view, nil
}

func (h *DetailsHandler) loadDetails(
	ctx context.Context,
	view *detailsView,
	appKey string,
) error {
	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			a.key,
			a.name,
			t.package,
			t.name,
			a.is_pointer,
			t.url,
			t.docs
		FROM docserve.application AS a
		INNER JOIN docserve.type AS t
		ON t.id = a.type_id
		WHERE a.key = $1`,
		appKey,
	)

	return row.Scan(
		&view.Key,
		&view.Name,
		&view.Impl.Package,
		&view.Impl.Name,
		&view.Impl.IsPointer,
		&view.Impl.URL,
		&view.Impl.Docs,
	)
}

func (h *DetailsHandler) loadHandlers(
	ctx context.Context,
	view *detailsView,
	appKey string,
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
			(
				SELECT COUNT(DISTINCT m.type_id)
				FROM docserve.handler_message AS m
				WHERE m.handler_key = h.key
			) AS message_count
		FROM docserve.handler AS h
		INNER JOIN docserve.type AS t
		ON t.id = h.type_id
		WHERE h.application_key = $1
		ORDER BY h.name`,
		appKey,
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
			&s.MessageCount,
		); err != nil {
			return err
		}

		view.Handlers = append(view.Handlers, s)
	}

	return rows.Err()
}

func (h *DetailsHandler) loadMessages(
	ctx context.Context,
	view *detailsView,
	appKey string,
) error {
	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT DISTINCT ON (t.name, t.package)
			t.package,
			t.name,
			m.is_pointer,
			t.url,
			t.docs,
			m.role,
			(
				SELECT COUNT(DISTINCT pm.handler_key)
				FROM docserve.handler_message AS pm
				WHERE pm.type_id = m.type_id
				AND pm.is_produced = TRUE
			) AS produced_count,
			(
				SELECT COUNT(DISTINCT pm.handler_key)
				FROM docserve.handler_message AS pm
				WHERE pm.type_id = m.type_id
				AND pm.is_consumed = TRUE
			) AS consumed_count
		FROM docserve.handler_message AS m
		INNER JOIN docserve.handler AS h
		ON h.key = m.handler_key
		INNER JOIN docserve.type AS t
		ON t.id = m.type_id
		WHERE h.application_key = $1
		ORDER BY t.name, t.package`,
		appKey,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var s messageSummary

		if err := rows.Scan(
			&s.Impl.Package,
			&s.Impl.Name,
			&s.Impl.IsPointer,
			&s.Impl.URL,
			&s.Impl.Docs,
			&s.Role,
			&s.ProducerCount,
			&s.ConsumerCount,
		); err != nil {
			return err
		}

		view.Messages = append(view.Messages, s)
	}

	return rows.Err()
}
