package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

func randomToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s Server) homePage(w http.ResponseWriter, r *http.Request) {
	user, exists := UserFromCtx(r.Context())
	if !exists {
		s.render500(w, fmt.Errorf("user not found"))
		return
	}
	s.render(w, 200, "home.tmpl.html", "", map[string]interface{}{
		"user":  user,
		"pages": pages,
	})
}

func (s Server) statusPage(w http.ResponseWriter, r *http.Request) {
	user, exists := UserFromCtx(r.Context())
	if !exists {
		s.render500(w, fmt.Errorf("user not found"))
		return
	}
	s.render(w, 200, "status.tmpl.html", "", map[string]interface{}{
		"user":  user,
		"pages": pages,
	})
}

func (s Server) auditPage(w http.ResponseWriter, r *http.Request) {
	user, exists := UserFromCtx(r.Context())
	if !exists {
		s.render500(w, fmt.Errorf("user not found"))
		return
	}
	logs, err := s.store.GetLogs(r.Context(), user.Id)
	if err != nil {
		s.render500(w, err)
		return
	}

	formatted := make([]struct {
		DateTime string
		Action   string
	}, len(logs))
	for idx, l := range logs {
		formatted[idx] = struct {
			DateTime string
			Action   string
		}{
			DateTime: l.CreatedAt.Format(time.DateTime),
			Action:   l.Action,
		}
	}

	s.render(w, 200, "audit.tmpl.html", "", map[string]interface{}{
		"user":  user,
		"logs":  formatted,
		"pages": pages,
	})
}

func (s Server) auth(w http.ResponseWriter, r *http.Request) {
	authToken, err := randomToken()
	if err != nil {
		s.render500(w, err)
		return
	}

	oneTimeToken := r.URL.Query().Get("token")

	user, err := s.store.ResolveOneTimeToken(r.Context(), oneTimeToken, authToken)
	if err != nil {
		s.render500(w, err)
		return
	}
	if user == nil {
		s.render500(w, fmt.Errorf("one time token [%s] has not been found", oneTimeToken))
		return
	}

	s.authenticator.Authenticate(r.Context(), w, user)
	http.Redirect(w, r, RouteHome, http.StatusSeeOther)
}

func (s Server) authRequired(w http.ResponseWriter, r *http.Request) {
	s.render(w, 200, "auth-required.tmpl.html", "", map[string]interface{}{
		"pages": pages,
	})
}
