package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"edu-portal/app"

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
	return nil
}

func (a *Api) Activate(w http.ResponseWriter, r *http.Request) {
	var req ActivateRequest
	if err := render.Bind(r, &req); err != nil {
		log.Printf("[ERROR] Bind req error: %v", err)
		w.WriteHeader(500)
		return
	}

	userIdRaw := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(userIdRaw)
	if err != nil {
		w.WriteHeader(400)
		render.Render(w, r, &ActivateResponse{
			Error: pointer.To("incorrect user id"),
		})
		return
	}

	user, err := a.Store.GetUserById(r.Context(), userId)
	if err != nil {
		log.Printf("[ERROR] Get user by id: %v", err)
		w.WriteHeader(500)
		return
	}
	if user == nil {
		w.WriteHeader(400)
		render.Render(w, r, &ActivateResponse{
			Error: pointer.To("user not found"),
		})
		return
	}

	defer func() {
		if err == nil {
			err = a.Store.Log(r.Context(), userId, "k8s user has been created")
		} else {
			err = a.Store.Log(r.Context(), userId, fmt.Sprintf("k8s user creation failed: %v", err))
		}

		if err != nil {
			log.Printf("[ERROR] Save audit log: %v", err)
		}
	}()

	username := fmt.Sprintf("edu-user-%d", user.Id)

	certificate, privateKey, err := a.Cluster.GenerateUserCertificate(r.Context(), username)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("[ERROR] Create cluster role binding: %v", err)
		return
	}

	if err := a.Store.SaveUserCertificate(r.Context(), &app.UserCertificate{
		UserId:      user.Id,
		Username:    username,
		Certificate: base64.StdEncoding.EncodeToString(certificate),
		PrivateKey:  base64.StdEncoding.EncodeToString(privateKey),
		CreatedAt:   time.Now(),
	}); err != nil {
		w.WriteHeader(500)
		log.Printf("[ERROR] Save user certificate: %v", err)
		return
	}

	if err := a.Cluster.CreateClusterRoleBinding(r.Context(), username); err != nil {
		w.WriteHeader(500)
		log.Printf("[ERROR] Create cluster role binding: %v", err)
		return
	}

	render.Render(w, r, &ActivateResponse{})
}
