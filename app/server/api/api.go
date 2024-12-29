package api

import (
	"context"
	"edu-portal/app"
)

type Store interface {
	AuthByOneTimeToken(ctx context.Context, authToken, oneTimeToken string) (bool, error)
	GetUserById(ctx context.Context, id int) (*app.User, error)
}

type Api struct {
	Store Store
}
