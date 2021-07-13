package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/dogmatiq/configkit"
	"github.com/google/go-github/v35/github"
)

type TypeDef struct {
	Package string
	Name    string
	File    string
	Line    int
}

func RepositoryNeedsSync(
	ctx context.Context,
	db *sql.DB,
	r *github.Repository,
	commit string,
) (bool, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT NOT EXISTS (
			SELECT *
			FROM docserve.repository
			WHERE id = $1
			AND commit_hash = $2
		)`,
		r.GetID(),
		commit,
	)

	var ok bool
	return ok, row.Scan(&ok)
}

func RemoveRepository(
	ctx context.Context,
	tx *sql.Tx,
	r *github.Repository,
) error {
	_, err := tx.ExecContext(
		ctx,
		`DELETE FROM docserve.repository
		WHERE github_id = $1`,
		r.GetID(),
	)
	return err
}

func SyncRepository(
	ctx context.Context,
	db *sql.DB,
	r *github.Repository,
	commit string,
	apps []configkit.Application,
	defs []TypeDef,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO docserve.repository AS r (
			id,
			full_name,
			commit_hash
		) VALUES (
			$1, $2, $3
		) ON CONFLICT (id) DO UPDATE SET
			full_name = excluded.full_name,
			commit_hash = excluded.commit_hash`,
		r.GetID(),
		r.GetFullName(),
		commit,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE docserve.application SET
			is_historical = TRUE
		WHERE repository_id = $1`,
		r.GetID(),
	); err != nil {
		return err
	}

	for _, a := range apps {
		if err := syncApplication(ctx, tx, r, a); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE docserve.type SET
			url = NULL
		WHERE repository_id = $1`,
		r.GetID(),
	); err != nil {
		return err
	}

	for _, t := range defs {
		if err := syncTypeDef(
			ctx,
			tx,
			r,
			t,
			commit,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func syncApplication(
	ctx context.Context,
	tx *sql.Tx,
	r *github.Repository,
	a configkit.Application,
) error {
	typeID, isPointer, err := syncType(ctx, tx, a.TypeName())
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO docserve.application (
			key,
			name,
			type_id,
			is_pointer,
			repository_id
		) VALUES (
			$1, $2, $3, $4, $5
		) ON CONFLICT (key) DO UPDATE SET
			name = excluded.name,
			type_id = excluded.type_id,
			is_pointer = excluded.is_pointer,
			repository_id = excluded.repository_id,
			is_historical = FALSE`,
		a.Identity().Key,
		a.Identity().Name,
		typeID,
		isPointer,
		r.GetID(),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE docserve.handler SET
			is_historical = TRUE
		WHERE application_key = $1`,
		a.Identity().Key,
	); err != nil {
		return err
	}

	for _, h := range a.Handlers() {
		if err := syncHandler(
			ctx,
			tx,
			a.Identity().Key,
			h,
		); err != nil {
			return err
		}
	}

	return nil
}

func syncHandler(
	ctx context.Context,
	tx *sql.Tx,
	appKey string,
	h configkit.Handler,
) error {
	typeID, isPointer, err := syncType(ctx, tx, h.TypeName())
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO docserve.handler (
			key,
			name,
			application_key,
			handler_type,
			type_id,
			is_pointer
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) ON CONFLICT (key) DO UPDATE SET
			name = excluded.name,
			application_key = excluded.application_key,
			handler_type = excluded.handler_type,
			type_id = excluded.type_id,
			is_pointer = excluded.is_pointer,
			is_historical = FALSE`,
		h.Identity().Key,
		h.Identity().Name,
		appKey,
		h.HandlerType(),
		typeID,
		isPointer,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE docserve.handler_message SET
			is_historical = TRUE
		WHERE handler_key = $1`,
		h.Identity().Key,
	); err != nil {
		return err
	}

	for n, r := range h.MessageNames().Produced {
		typeID, isPointer, err := syncType(ctx, tx, n.String())
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO docserve.handler_message (
				handler_key,
				type_id,
				is_pointer,
				role,
				is_produced
			) VALUES (
				$1, $2, $3, $4, TRUE
			) ON CONFLICT (handler_key, type_id, is_pointer) DO UPDATE SET
				role = excluded.role,
				is_produced = excluded.is_produced,
				is_historical = FALSE`,
			h.Identity().Key,
			typeID,
			isPointer,
			r,
		); err != nil {
			return err
		}
	}

	for n, r := range h.MessageNames().Consumed {
		typeID, isPointer, err := syncType(ctx, tx, n.String())
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO docserve.handler_message (
				handler_key,
				type_id,
				is_pointer,
				role,
				is_consumed
			) VALUES (
				$1, $2, $3, $4, TRUE
			) ON CONFLICT (handler_key, type_id, is_pointer) DO UPDATE SET
				role = excluded.role,
				is_consumed = excluded.is_consumed,
				is_historical = FALSE`,
			h.Identity().Key,
			typeID,
			isPointer,
			r,
		); err != nil {
			return err
		}
	}

	return nil
}

func syncType(
	ctx context.Context,
	tx *sql.Tx,
	name string,
) (typeID int, isPointer bool, err error) {
	if name[0] == '*' {
		isPointer = true
		name = name[1:]
	}

	var pkg string
	if n := strings.LastIndexByte(name, '.'); n != -1 {
		pkg = name[:n]
		name = name[n+1:]
	}

	row := tx.QueryRowContext(
		ctx,
		`INSERT INTO docserve.type (
			package,
			name
		) VALUES (
			$1, $2
		) ON CONFLICT (package, name) DO UPDATE SET
			package = excluded.package
		RETURNING id`, // DO UPDATE is a no-op that allows use of RETURNING when row already exists
		pkg,
		name,
	)

	if err := row.Scan(&typeID); err != nil {
		return 0, false, err
	}

	return typeID, isPointer, nil
}

func syncTypeDef(
	ctx context.Context,
	tx *sql.Tx,
	r *github.Repository,
	t TypeDef,
	commit string,
) error {
	us := r.GetHTMLURL()

	u, err := url.Parse(us)
	if err != nil {
		return err
	}

	u.Path = path.Join(
		u.Path,
		"blob",
		commit,
		t.File,
	)

	u.Fragment = fmt.Sprintf("L%d", t.Line)

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO docserve.type (
			package,
			name,
			repository_id,
			url
		) VALUES (
			$1, $2, $3, $4
		) ON CONFLICT (package, name) DO UPDATE SET
			repository_id = excluded.repository_id,
			url = excluded.url`,
		t.Package,
		t.Name,
		r.GetID(),
		u.String(),
	)

	return err
}
