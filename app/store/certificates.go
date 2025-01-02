package store

import (
	"context"
	"database/sql"
	"edu-portal/app"
	"errors"
)

func (s *Store) SaveUserCertificate(ctx context.Context, certificate *app.UserCertificate) error {
	query := `
		INSERT INTO k8s_certificates (user_id, username, certificate, private_key, created_at)
		VALUES (:user_id, :username, :certificate, :private_key, :created_at)
	`
	_, err := s.db.NamedExecContext(ctx, query, certificate)
	return err
}

func (s *Store) GetUserCertificate(ctx context.Context, userId int) (*app.UserCertificate, error) {
	query := `SELECT * FROM k8s_certificates WHERE user_id = ?`

	var cert app.UserCertificate

	if err := s.db.QueryRowxContext(ctx, query, userId).StructScan(&cert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &cert, nil
}
