package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/docserve/web/components"
	"github.com/gin-gonic/gin"
)

type detailsView struct {
	Key     string
	Name    string
	Type    configkit.HandlerType
	Impl    components.Type
	AppKey  string
	AppName string

	ConsumedMessages    []components.Type
	ConsumedMessageRole message.Role
	ProducedMessages    []components.Type
	ProducedMessageRole message.Role
	TimeoutMessages     []components.Type
}

type DetailsHandler struct {
	DB *sql.DB
}

func (h *DetailsHandler) Route() (string, string) {
	return http.MethodGet, "/handlers/:key"
}

func (h *DetailsHandler) Template() string {
	return "handlers/details.html"
}

func (h *DetailsHandler) ActiveMenuItem() components.MenuItem {
	return components.HandlersMenuItem
}

func (h *DetailsHandler) View(ctx *gin.Context) (string, interface{}, error) {
	var view detailsView

	handlerKey := ctx.Param("key")

	if err := h.loadDetails(ctx, &view, handlerKey); err != nil {
		if err == sql.ErrNoRows {
			ctx.AbortWithStatus(http.StatusNotFound)
			return "", nil, nil
		}

		return "", nil, err
	}

	switch view.Type {
	case configkit.AggregateHandlerType, configkit.IntegrationHandlerType:
		view.ConsumedMessageRole = message.CommandRole
		view.ProducedMessageRole = message.EventRole
	case configkit.ProcessHandlerType:
		view.ConsumedMessageRole = message.EventRole
		view.ProducedMessageRole = message.CommandRole
	case configkit.ProjectionHandlerType:
		view.ConsumedMessageRole = message.EventRole
	}

	if err := h.loadMessages(ctx, &view, handlerKey); err != nil {
		return "", nil, err
	}

	return view.Name + " - Handler Details", view, nil
}

func (h *DetailsHandler) loadDetails(
	ctx context.Context,
	view *detailsView,
	handlerKey string,
) error {
	row := h.DB.QueryRowContext(
		ctx,
		`SELECT
			h.key,
			h.name,
			h.handler_type,
			t.package,
			t.name,
			h.is_pointer,
			COALESCE(t.url, ''),
			COALESCE(t.docs, ''),
			a.key,
			a.name
		FROM docserve.handler AS h
		INNER JOIN docserve.type AS t
		ON t.id = h.type_id
		INNER JOIN docserve.application AS a
		ON a.key = h.application_key
		WHERE h.key = $1`,
		handlerKey,
	)

	return row.Scan(
		&view.Key,
		&view.Name,
		&view.Type,
		&view.Impl.Package,
		&view.Impl.Name,
		&view.Impl.IsPointer,
		&view.Impl.URL,
		&view.Impl.Docs,
		&view.AppKey,
		&view.AppName,
	)
}

func (h *DetailsHandler) loadMessages(
	ctx context.Context,
	view *detailsView,
	handlerKey string,
) error {
	rows, err := h.DB.QueryContext(
		ctx,
		`SELECT
			t.package,
			t.name,
			m.is_pointer,
			COALESCE(t.url, ''),
			COALESCE(t.docs, ''),
			m.role,
			m.is_produced,
			m.is_consumed
		FROM docserve.handler_message AS m
		INNER JOIN docserve.type AS t
		ON t.id = m.type_id
		WHERE m.handler_key = $1
		ORDER BY t.name, t.package`,
		handlerKey,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var t components.Type

		var (
			role                   message.Role
			isProduced, isConsumed bool
		)

		if err := rows.Scan(
			&t.Package,
			&t.Name,
			&t.IsPointer,
			&t.URL,
			&t.Docs,
			&role,
			&isProduced,
			&isConsumed,
		); err != nil {
			return err
		}

		if role == message.TimeoutRole {
			view.TimeoutMessages = append(view.ProducedMessages, t)
		} else if isProduced {
			view.ProducedMessages = append(view.ProducedMessages, t)
		} else if isConsumed {
			view.ConsumedMessages = append(view.ConsumedMessages, t)
		}
	}

	return rows.Err()
}
