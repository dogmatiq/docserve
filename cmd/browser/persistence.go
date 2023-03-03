package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/dogmatiq/browser/persistence"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	provide(func() (*sql.DB, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		db, err := sql.Open("pgx", postgresDSN.Value())
		if err != nil {
			return nil, err
		}

		db.SetConnMaxIdleTime(30 * time.Second)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetMaxIdleConns(3)
		db.SetMaxOpenConns(50)

		if err := persistence.CreateSchema(ctx, db); err != nil {
			return nil, err
		}

		return db, nil
	})
}
