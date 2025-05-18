package usecases

import (
	"edu-portal/app"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	IsStaff bool `json:"is_staff"`
	jwt.RegisteredClaims
}

type GenerateJWT struct {
	privateKey []byte
	publicKey  []byte
}

func NewGenerateJWT(privateKey []byte, publicKey []byte) *GenerateJWT {
	return &GenerateJWT{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (g *GenerateJWT) Token(user *app.User) (string, error) {
	privateKeyParsed, err := jwt.ParseRSAPrivateKeyFromPEM(g.privateKey)
	if err != nil {
		return "", err
	}
	claims := CustomClaims{
		IsStaff: user.IsStaff,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(user.Id),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKeyParsed)
}

func (g *GenerateJWT) PublicKey() string {
	return string(g.publicKey)
}
