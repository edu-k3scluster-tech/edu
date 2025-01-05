package api

import (
	"fmt"
	"net/http"
	"strconv"

	"edu-portal/app/server/utils"

	"github.com/AlekSi/pointer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ActivateRequest struct{}

func (a *ActivateRequest) Bind(r *http.Request) error {
	return nil
}

type ActivateResponse struct {
	Error *string `json:"error"`
}

func (a *ActivateResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if a.Error != nil {
		w.WriteHeader(400)
	}
	return nil
}

// method to activate a user
func (a *Api) Activate(w http.ResponseWriter, r *http.Request) {
	var req ActivateRequest
	if err := render.Bind(r, &req); err != nil {
		render.Render(w, r, &ActivateResponse{
			Error: pointer.To("invalid request"),
		})
		return
	}

	userIdRaw := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(userIdRaw)
	if err != nil {
		render.Render(w, r, &ActivateResponse{Error: pointer.To("incorrect user id")})
		return
	}

	user, err := a.Store.GetUserById(r.Context(), userId)
	if err != nil {
		utils.Render500(w, fmt.Errorf("get user by id: %w", err))
		return
	}
	if user == nil {
		render.Render(w, r, &ActivateResponse{Error: pointer.To("user not found")})
		return
	}

	if err := a.ActivateUC.Do(r.Context(), user); err != nil {
		utils.Render500(w, fmt.Errorf("activate uc: %w", err))
		return
	}

	render.Render(w, r, &ActivateResponse{})
}
