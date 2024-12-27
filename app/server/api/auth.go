package api

import (
	"edu-portal/pkg/token"
	"log"
	"net/http"

	"github.com/go-chi/render"
)

type AuthRequest struct {
	OneTimeToken string `json:"one_time_token"`
}

func (a *AuthRequest) Bind(r *http.Request) error {
	return nil
}

type AuthResponse struct {
	AuthToken *string `json:"auth_token"`
}

func (a *AuthResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *Api) Auth(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := render.Bind(r, &req); err != nil {
		log.Printf("[ERROR] Bind req error: %v", err)
		render.Render(w, r, &AuthResponse{})
		return
	}

	authToken := token.RandomToken()
	created, err := a.Store.AuthByOneTimeToken(r.Context(), authToken, req.OneTimeToken)
	if err != nil {
		log.Printf("[ERROR] Auth by one time token: %v", err)
		render.Render(w, r, &AuthResponse{})
		return
	}

	if !created {
		render.Render(w, r, &AuthResponse{AuthToken: nil})
		return
	}

	render.Render(w, r, &AuthResponse{AuthToken: &authToken})
}
