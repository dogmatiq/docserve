package persistence

import (
	"context"
	"database/sql"

	"github.com/dogmatiq/configkit"
	"github.com/google/go-github/v35/github"
)

func SyncApplication(
	ctx context.Context,
	tx *sql.Tx,
	r *github.Repository,
	cfg configkit.Application,
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
			repository_id = excluded.repository_id`,
		cfg.Identity().Key,
		cfg.Identity().Name,
		cfg.TypeName(),
		r.GetID(),
	); err != nil {
		return err
	}

	for _, hcfg := range cfg.Handlers() {
		if err := syncHandler(
			ctx,
			tx,
			cfg.Identity().Key,
			hcfg,
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
	cfg configkit.Handler,
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
			type_name = excluded.type_name`,
		cfg.Identity().Key,
		cfg.Identity().Name,
		appKey,
		cfg.HandlerType(),
		cfg.TypeName(),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM docserve.handler_message
		WHERE handler_key = $1`,
		cfg.Identity().Key,
	); err != nil {
		return err
	}

	for n, r := range cfg.MessageNames().Produced {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO docserve.handler_message (
				handler_key,
				type_name,
				role,
				produced
			) VALUES (
				$1, $2, $3, true
			) ON CONFLICT (handler_key, type_name) DO UPDATE SET
				role = excluded.role,
				produced = excluded.produced`,
			cfg.Identity().Key,
			n.String(),
			r,
		); err != nil {
			return err
		}
	}

	for n, r := range cfg.MessageNames().Consumed {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO docserve.handler_message (
				handler_key,
				type_name,
				role,
				consumed
			) VALUES (
				$1, $2, $3, true
			) ON CONFLICT (handler_key, type_name) DO UPDATE SET
				role = excluded.role,
				consumed = excluded.consumed`,
			cfg.Identity().Key,
			n.String(),
			r,
		); err != nil {
			return err
		}
	}

	return nil
}
