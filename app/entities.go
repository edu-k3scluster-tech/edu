package app

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"
)

type UserStatus string

const (
	UserStatusNew         UserStatus = "new"
	UserStatusActive      UserStatus = "active"
	UserStatusDeactivated UserStatus = "deactivated"
)

type User struct {
	Id         int        `db:"id"`
	TgId       *int64     `db:"tg_id"`
	TgUsername *string    `db:"tg_username"`
	Status     UserStatus `db:"status"`
	IsStaff    bool       `db:"is_staff"`
	CreatedAt  time.Time  `db:"created_at"`
}

type AuditLog struct {
	UserId    int       `db:"user_id"`
	Action    string    `db:"action"`
	CreatedAt time.Time `db:"created_at"`
}

type AuthToken struct {
	UserId    int       `db:"user_id"`
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"created_at"`
}

type TgOneTimeToken struct {
	Token     string    `db:"token"`
	UserId    *int      `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type UserCertificate struct {
	UserId             int       `db:"user_id"`
	Username           string    `db:"username"`
	Certificate        *string   `db:"certificate"`
	CertificateRequest string    `db:"certificate_request"`
	PrivateKey         string    `db:"private_key"`
	CreatedAt          time.Time `db:"created_at"`
}

func (c *UserCertificate) SetCertificate(certData []byte) {
	c.Certificate = pointer.To(base64.StdEncoding.EncodeToString(certData))
}

func (c *UserCertificate) GetCertificate() ([]byte, error) {
	if cdata := pointer.Get(c.Certificate); cdata == "" {
		return nil, nil
	} else {
		return base64.StdEncoding.DecodeString(*c.Certificate)
	}
}

func (c *UserCertificate) SetPrivateKey(key *rsa.PrivateKey) {
	c.PrivateKey = base64.StdEncoding.EncodeToString(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			},
		),
	)
}

func (c *UserCertificate) GetPrivateKey() (*rsa.PrivateKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(c.PrivateKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(decoded)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing RSA private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}
