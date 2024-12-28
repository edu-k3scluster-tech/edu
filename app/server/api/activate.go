package api

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
)

type ActivateRequest struct {
	Id int `json:"user_id"`
}

func (a *ActivateRequest) Bind(r *http.Request) error {
	return nil
}

type ActivateResponse struct{}

func (a *ActivateResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *Api) Activate(w http.ResponseWriter, r *http.Request) {
	var req ActivateRequest
	if err := render.Bind(r, &req); err != nil {
		log.Printf("[ERROR] Bind req error: %v", err)
		render.Render(w, r, &ActivateResponse{})
		return
	}

	// config, err := createuser.New().Create()
	render.Render(w, r, &ActivateResponse{})
}
