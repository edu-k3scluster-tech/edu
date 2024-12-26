package migrator

import (
	"context"
	"embed"
	"io/fs"

	"github.com/jmoiron/sqlx"
)

//go:embed "migrations"
var Migrations embed.FS

type Migrator struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Migrator {
	return &Migrator{
		db: db,
	}
}

func (m *Migrator) Run(ctx context.Context) error {
	files, err := fs.Glob(Migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	for _, f := range files {
		q, err := Migrations.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := m.db.ExecContext(ctx, string(q)); err != nil {
			return err
		}
	}
	return nil
}
