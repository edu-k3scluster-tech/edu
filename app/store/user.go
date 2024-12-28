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
		INSERT INTO users (tg_id, tg_username, status, is_staff, created_at)
		VALUES (:tg_id, :tg_username, :status, :is_staff, :created_at)
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

func (s *Store) GrantAdminPermissions(ctx context.Context, userId int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	query := `
		UPDATE users
		SET is_staff = true, status = 'active'
		WHERE id = ?
	`
	result, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		if err = s.logWithTx(ctx, tx, userId, "admin permissions have been granted"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetUserById(ctx context.Context, id int) (*app.User, error) {
	query := `SELECT * FROM users WHERE id = ?`

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

func (s *Store) GetUsers(ctx context.Context) ([]app.User, error) {
	query := `
		SELECT *
		FROM users
	`
	var users []app.User
	if err := s.db.SelectContext(ctx, &users, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return users, nil
}
