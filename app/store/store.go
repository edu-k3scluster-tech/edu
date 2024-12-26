package store

import (
	"context"
	"database/sql"
	"edu-portal/app"
	"errors"
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

func (s *Store) ResolveOneTimeToken(ctx context.Context, oneTimeToken, authToken string) (*app.User, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var t app.OneTimeToken
	if err := tx.QueryRowxContext(ctx, "SELECT * FROM one_time_tokens WHERE token = ?", oneTimeToken).StructScan(&t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM one_time_tokens WHERE token = ?", oneTimeToken); err != nil {
		return nil, err
	}

	var u app.User
	if err := tx.QueryRowxContext(ctx, "UPDATE users SET auth_token = ? WHERE id = ? RETURNING id, tg_id, auth_token", authToken, t.UserId).StructScan(&u); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if err := s.log(ctx, tx, u.Id, "authentication by one-time token succeeded"); err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *Store) GetUserByAuthToken(ctx context.Context, token string) (*app.User, error) {
	var user app.User
	if err := s.db.QueryRowxContext(ctx, "SELECT * FROM users WHERE auth_token = ?", token).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &user, nil
}

func (s *Store) GetLogs(ctx context.Context, userId string) ([]app.AuditLog, error) {
	var logs []app.AuditLog
	if err := s.db.SelectContext(ctx, &logs, "SELECT * FROM audit_logs WHERE user_id = ?", userId); err != nil {
		return nil, err
	} else {
		return logs, nil
	}
}

func (s *Store) log(ctx context.Context, tx *sqlx.Tx, userId, action string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO audit_logs (user_id, action, created_at) VALUES (?, ?, ?)", userId, action, time.Now())
	return err
}
