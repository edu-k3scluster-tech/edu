package server

import (
	"context"
	"edu-portal/app"
	"errors"
	"net/http"
)

const sessionTokenParam = "session_token"

type Authenticator struct {
	secured  bool
	resolver func(context.Context, string) (*app.User, error)
}

func (a *Authenticator) IsAuthenticated(ctx context.Context, r *http.Request) (*app.User, error) {
	cookie, err := r.Cookie(sessionTokenParam)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	user, err := a.resolver(r.Context(), cookie.Value)
	if err != nil {
		return nil, err
	}
	return user, err
}

func (a *Authenticator) Authenticate(ctx context.Context, w http.ResponseWriter, user *app.User) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenParam,
		Value:    user.Token,
		HttpOnly: true,
		Secure:   a.secured,
		Path:     "/",
	})
}
