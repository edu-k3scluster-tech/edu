package store

import (
	"context"
	"database/sql"
	"edu-portal/app"
	"errors"
	"fmt"
)

func (s *Store) CreateUser(ctx context.Context, user *app.User) error {
	query := `
		INSERT INTO users (tg_id, tg_username)
		VALUES (:tg_id, :tg_username)
		RETURNING id;
	`
	rows, err := s.db.NamedQuery(query, user)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&user.Id); err != nil {
			return fmt.Errorf("retrieve new id: %w", err)
		}
	}

	return nil
}

func (s *Store) GetUserByTgId(ctx context.Context, id int64) (*app.User, error) {
	query := `SELECT * FROM users WHERE tg_id = ?`

	var user app.User

	if err := s.db.QueryRowxContext(ctx, query, id).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &user, nil
}

func (s *Store) GetUserByAuthToken(ctx context.Context, token string) (*app.User, error) {
	query := `
		SELECT u.*
		FROM users u
		LEFT JOIN auth_tokens t ON t.user_id = u.id
		WHERE t.token = ?
	`

	var user app.User
	if err := s.db.QueryRowxContext(ctx, query, token).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &user, nil
}