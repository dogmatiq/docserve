package persistence

import (
	"context"
	"database/sql"

	"github.com/google/go-github/v35/github"
)

func SyncRepository(
	ctx context.Context,
	tx *sql.Tx,
	r *github.Repository,
	commitHash string,
) error {
	_, err := tx.ExecContext(
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
	)

	return err
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
