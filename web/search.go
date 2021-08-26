package web

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type searchTerm struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Docs string `json:"docs,omitempty"`
	URI  string `json:"uri"`
}

func searchTerms(version string, db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var terms []searchTerm

		t, err := applicationSearchTerms(ctx, db)
		if err != nil {
			fmt.Println(err) // TODO
			renderError(ctx, version, http.StatusInternalServerError)
			return
		}
		terms = append(terms, t...)

		t, err = handlerSearchTerms(ctx, db)
		if err != nil {
			fmt.Println(err) // TODO
			renderError(ctx, version, http.StatusInternalServerError)
			return
		}
		terms = append(terms, t...)

		t, err = messageSearchTerms(ctx, db)
		if err != nil {
			fmt.Println(err) // TODO
			renderError(ctx, version, http.StatusInternalServerError)
			return
		}
		terms = append(terms, t...)

		ctx.PureJSON(http.StatusOK, terms)
	}
}

func applicationSearchTerms(ctx context.Context, db *sql.DB) ([]searchTerm, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT
			a.key,
			a.name,
			t.docs
		FROM dogmabrowser.application AS a
		INNER JOIN dogmabrowser.type AS t
		ON t.id = a.type_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to query applications: %w", err)
	}

	var terms []searchTerm

	for rows.Next() {
		var (
			key string
			t   searchTerm
		)

		if err := rows.Scan(
			&key,
			&t.Name,
			&t.Docs,
		); err != nil {
			return nil, fmt.Errorf("unable scan application result: %w", err)
		}

		t.Type = "application"
		t.URI = "/applications/" + key

		terms = append(terms, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unable to iterate all application rows: %w", err)
	}

	return terms, nil
}

func handlerSearchTerms(ctx context.Context, db *sql.DB) ([]searchTerm, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT
			h.key,
			h.name,
			h.handler_type,
			t.docs
		FROM dogmabrowser.handler AS h
		INNER JOIN dogmabrowser.type AS t
		ON t.id = h.type_id
		INNER JOIN dogmabrowser.application AS a
		ON a.key = h.application_key`,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to query handlers: %w", err)
	}

	var terms []searchTerm

	for rows.Next() {
		var (
			key string
			t   searchTerm
		)

		if err := rows.Scan(
			&key,
			&t.Name,
			&t.Type,
			&t.Docs,
		); err != nil {
			return nil, fmt.Errorf("unable scan handler result: %w", err)
		}

		t.URI = "/handlers/" + key

		terms = append(terms, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unable to iterate all handler rows: %w", err)
	}

	return terms, nil
}

func messageSearchTerms(ctx context.Context, db *sql.DB) ([]searchTerm, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT
			t.package,
			t.name,
			MODE() WITHIN GROUP (ORDER BY m.role),
			t.docs
		FROM dogmabrowser.handler_message AS m
		INNER JOIN dogmabrowser.type AS t
		ON t.id = m.type_id
		GROUP BY t.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to query handlers: %w", err)
	}

	var terms []searchTerm

	for rows.Next() {
		var (
			pkg string
			t   searchTerm
		)

		if err := rows.Scan(
			&pkg,
			&t.Name,
			&t.Type,
			&t.Docs,
		); err != nil {
			return nil, fmt.Errorf("unable scan handler result: %w", err)
		}

		t.URI = "/messages/" + pkg + "." + t.Name

		terms = append(terms, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unable to iterate all handler rows: %w", err)
	}

	return terms, nil
}
