package store

import (
	"context"
	"edu-portal/app"
	"time"

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

func (s *Store) logWithTx(ctx context.Context, tx *sqlx.Tx, userId int, msg string) error {
	query := `
		INSERT INTO audit_logs (user_id, action, created_at)
		VALUES (:user_id, :action, :created_at)
	`
	_, err := tx.NamedExecContext(ctx, query, app.AuditLog{
		UserId:    userId,
		Action:    msg,
		CreatedAt: time.Now(),
	})
	return err
}

func (s *Store) Log(ctx context.Context, userId int, msg string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Commit()
	return s.logWithTx(ctx, tx, userId, msg)
}
