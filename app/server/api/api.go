package api

import (
	"context"
	"edu-portal/app"
)

type Store interface {
	AuthByOneTimeToken(ctx context.Context, authToken, oneTimeToken string) (bool, error)
	GetUserById(ctx context.Context, id int) (*app.User, error)
	Log(ctx context.Context, userId int, msg string) error
}

type Api struct {
	Store Store
}
