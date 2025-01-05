package api

import (
	"context"
	"edu-portal/app"
)

type Store interface {
	AuthByOneTimeToken(ctx context.Context, authToken, oneTimeToken string) (bool, error)
	GetUserById(ctx context.Context, id int) (*app.User, error)
	Log(ctx context.Context, userId int, msg string) error
	GetUserCertificate(ctx context.Context, userId int) (*app.UserCertificate, error)
	SaveUserCertificate(ctx context.Context, certificate *app.UserCertificate) error
}

type Cluster interface {
	GenerateUserCertificate(ctx context.Context, username string) ([]byte, []byte, error)
	CreateClusterRoleBinding(ctx context.Context, username string) error
}

type Activate interface {
	Do(ctx context.Context, user *app.User) error
}

type Api struct {
	Store      Store
	Cluster    Cluster
	ActivateUC Activate
}
