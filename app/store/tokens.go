package store

import (
	"context"
	"database/sql"
	"edu-portal/app"
	"errors"
	"time"

	"github.com/AlekSi/pointer"
)

func (s *Store) AssignOneTimeToken(ctx context.Context, token string) error {
	query := `
		INSERT INTO tg_one_time_tokens (token, created_at)
		VALUES (?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, token, time.Now())
	return err
}

func (s *Store) ResolveTgToken(ctx context.Context, user_id int, token string) error {
	query := `
		UPDATE tg_one_time_tokens
		SET user_id = ?
		WHERE token = ?
	`
	_, err := s.db.ExecContext(ctx, query, user_id, token)
	return err
}

func (s *Store) AuthByOneTimeToken(ctx context.Context, authToken, oneTimeToken string) (bool, error) {
	query := `
		SELECT *
		FROM tg_one_time_tokens
		WHERE token = ? AND user_id IS NOT NULL
	`

	dbTgOneTimeToken := app.TgOneTimeToken{}
	if err := s.db.QueryRowxContext(ctx, query, oneTimeToken).StructScan(&dbTgOneTimeToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}
	if dbTgOneTimeToken.UserId == nil {
		return false, nil
	}

	dbAuthToken := app.AuthToken{
		UserId:    pointer.Get(dbTgOneTimeToken.UserId),
		Token:     authToken,
		CreatedAt: time.Now(),
	}

	query = `
		INSERT INTO auth_tokens (user_id, token, created_at)
		VALUES (:user_id, :token, :created_at)
	`
	if _, err := s.db.NamedExecContext(ctx, query, &dbAuthToken); err != nil {
		return false, err
	}
	return true, nil
}
