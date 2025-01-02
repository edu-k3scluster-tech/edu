package api

import (
	"context"
	"edu-portal/app"
)

type Store interface {
	AuthByOneTimeToken(ctx context.Context, authToken, oneTimeToken string) (bool, error)
	GetUserById(ctx context.Context, id int) (*app.User, error)
	Log(ctx context.Context, userId int, msg string) error
	SaveUserCertificate(ctx context.Context, certificate *app.UserCertificate) error
}

type Cluster interface {
	GenerateUserCertificate(ctx context.Context, username string) ([]byte, []byte, error)
	CreateClusterRoleBinding(ctx context.Context, username string) error
}

type Api struct {
	Store   Store
	Cluster Cluster
}
