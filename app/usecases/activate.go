package usecases

import (
	"context"
	"crypto/rsa"
	"edu-portal/app"
	"edu-portal/pkg/cert"
	"encoding/base64"
	"fmt"
	"log"
	"time"
)

type activateStore interface {
	GetUserById(ctx context.Context, id int) (*app.User, error)
	ChangeStatus(ctx context.Context, userId int, status string) error
	GetUserCertificate(ctx context.Context, userId int) (*app.UserCertificate, error)
	SaveUserCertificate(ctx context.Context, certificate *app.UserCertificate) error
}

type activateCluster interface {
	CertificateSigningRequest(ctx context.Context, username string, key *rsa.PrivateKey) error
	CreateClusterRoleBinding(ctx context.Context, username, role string) error
	CreateRoleBinding(ctx context.Context, username, namespace, role string) error
	Certificate(ctx context.Context, name string) ([]byte, bool, error)
}

type ActivateUser struct {
	store   activateStore
	cluster activateCluster
}

func NewActivateUser(store activateStore, cluster activateCluster) *ActivateUser {
	return &ActivateUser{
		store:   store,
		cluster: cluster,
	}
}

// method to activate a user. includes the following steps:
// 1. generating RSA key
// 2. making CSR request to k8s cluster (and approving it)
// 3. sabing certificate and private key to the db
// 4. changing the user's status to `active`
func (u *ActivateUser) Do(ctx context.Context, user *app.User) error {
	certificate, err := u.store.GetUserCertificate(ctx, user.Id)
	if err != nil {
		return err
	}

	if certificate == nil {
		certificate, err = u.generateAndSave(ctx, user)
		if err != nil {
			return err
		}
	}

	key, err := certificate.GetPrivateKey()
	if err != nil {
		return err
	}

	if err := u.cluster.CertificateSigningRequest(ctx, certificate.Username, key); err != nil {
		return err
	}

	clusterRoles := []string{
		"ro-user-role",
	}
	for _, role := range clusterRoles {
		if err := u.cluster.CreateClusterRoleBinding(ctx, certificate.Username, role); err != nil {
			return err
		}
	}

	namespaceToRole := map[string]string{
		"events-provider-staging": "ro-user-exec-portforward-role",
		"kafka":                   "ro-user-exec-portforward-role",
	}
	for ns, role := range namespaceToRole {
		if err := u.cluster.CreateRoleBinding(ctx, certificate.Username, ns, role); err != nil {
			return err
		}
	}

	if certData, ready, err := u.cluster.Certificate(ctx, certificate.Username); err != nil {
		return err
	} else {
		if !ready {
			log.Printf("[INFO] Certificate is not ready yet")
		} else {
			certificate.SetCertificate(certData)
		}
	}

	if err := u.store.SaveUserCertificate(ctx, certificate); err != nil {
		return err
	}

	if err := u.store.ChangeStatus(ctx, user.Id, "active"); err != nil {
		return err
	}

	return nil
}

func (u *ActivateUser) generateAndSave(ctx context.Context, user *app.User) (*app.UserCertificate, error) {
	username := fmt.Sprintf("edu-user-%d", user.Id)

	csr, key, err := cert.GeneratePrivateKeyAndCSR(username)
	if err != nil {
		return nil, err
	}

	certificate := &app.UserCertificate{
		UserId:             user.Id,
		Username:           username,
		Certificate:        nil,
		CertificateRequest: base64.StdEncoding.EncodeToString(csr),
		CreatedAt:          time.Now(),
	}
	certificate.SetPrivateKey(key)

	if err := u.store.SaveUserCertificate(ctx, certificate); err != nil {
		return nil, err
	}

	return certificate, nil
}
