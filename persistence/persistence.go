package persistence

import (
	"context"
	"database/sql"

	"github.com/dogmatiq/configkit"
	"github.com/google/go-github/v35/github"
)

func RepositoryNeedsSync(
	ctx context.Context,
	db *sql.DB,
	r *github.Repository,
	commitHash string,
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
		commitHash,
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
	commitHash string,
	apps []configkit.Application,
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
		commitHash,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE docserve.application SET
			historical = TRUE
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

	return tx.Commit()
}

func syncApplication(
	ctx context.Context,
	tx *sql.Tx,
	r *github.Repository,
	a configkit.Application,
) error {
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO docserve.application (
			key,
			name,
			type_name,
			repository_id
		) VALUES (
			$1, $2, $3, $4
		) ON CONFLICT (key) DO UPDATE SET
			name = excluded.name,
			type_name = excluded.type_name,
			repository_id = excluded.repository_id,
			historical = FALSE`,
		a.Identity().Key,
		a.Identity().Name,
		a.TypeName(),
		r.GetID(),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE docserve.handler SET
			historical = TRUE
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
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO docserve.handler (
			key,
			name,
			application_key,
			handler_type,
			type_name
		) VALUES (
			$1, $2, $3, $4, $5
		) ON CONFLICT (key) DO UPDATE SET
			name = excluded.name,
			application_key = excluded.application_key,
			handler_type = excluded.handler_type,
			type_name = excluded.type_name,
			historical = FALSE`,
		h.Identity().Key,
		h.Identity().Name,
		appKey,
		h.HandlerType(),
		h.TypeName(),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE docserve.handler_message SET
			historical = TRUE
		WHERE handler_key = $1`,
		h.Identity().Key,
	); err != nil {
		return err
	}

	for n, r := range h.MessageNames().Produced {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO docserve.handler_message (
				handler_key,
				type_name,
				role,
				produced
			) VALUES (
				$1, $2, $3, TRUE
			) ON CONFLICT (handler_key, type_name) DO UPDATE SET
				role = excluded.role,
				produced = excluded.produced,
				historical = FALSE`,
			h.Identity().Key,
			n.String(),
			r,
		); err != nil {
			return err
		}
	}

	for n, r := range h.MessageNames().Consumed {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO docserve.handler_message (
				handler_key,
				type_name,
				role,
				consumed
			) VALUES (
				$1, $2, $3, TRUE
			) ON CONFLICT (handler_key, type_name) DO UPDATE SET
				role = excluded.role,
				consumed = excluded.consumed,
				historical = FALSE`,
			h.Identity().Key,
			n.String(),
			r,
		); err != nil {
			return err
		}
	}

	return nil
}
