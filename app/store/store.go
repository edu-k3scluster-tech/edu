package store

import (
	"context"
	"edu-portal/app"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetLogs(ctx context.Context, userId int) ([]app.AuditLog, error) {
	query := `
		SELECT *
		FROM audit_logs
		WHERE user_id = ?
	`
	var logs []app.AuditLog
	if err := s.db.SelectContext(ctx, &logs, query, userId); err != nil {
		return nil, err
	} else {
		return logs, nil
	}
}
