package api

import "context"

type Store interface {
	AuthByOneTimeToken(ctx context.Context, authToken, oneTimeToken string) (bool, error)
}

type Api struct {
	Store Store
}
