package server

import (
	"bytes"
	"context"
	"edu-portal/app"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Store interface {
	GetUserByAuthToken(ctx context.Context, token string) (*app.User, error)
	ResolveOneTimeToken(ctx context.Context, oneTimeToken, authToken string) (*app.User, error)
	GetLogs(ctx context.Context, userId string) ([]app.AuditLog, error)
}

type Server struct {
	templates     map[string]*template.Template
	store         Store
	authenticator *Authenticator
}

func New(secured bool, templates map[string]*template.Template, store Store) *Server {
	return &Server{
		templates: templates,
		store:     store,
		authenticator: &Authenticator{
			secured:  secured,
			resolver: store.GetUserByAuthToken,
		},
	}
}

func (s Server) Run(ctx context.Context) error {
	return http.ListenAndServe(":8000", s.routes())

}

func (s Server) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.render500, s.authenticator))

		r.Get(RouteHome, s.homePage)
		r.Get(RouteStatus, s.statusPage)
		r.Get(RouteAudit, s.auditPage)
	})

	r.Get(RouteAuthRequired, s.authRequired)
	r.Get(RouteAuth, s.auth)

	return r
}

func (s Server) render(w http.ResponseWriter, status int, page, tmplName string, data any) {
	ts, ok := s.templates[page]
	if !ok {
		s.render500(w, fmt.Errorf("the template %s does not exist", page))
		return
	}

	buf := new(bytes.Buffer)

	if tmplName == "" {
		tmplName = "base"
	}

	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		s.render500(w, err)
		return
	}

	w.WriteHeader(status)
	if _, err = buf.WriteTo(w); err != nil {
		s.render500(w, err)
		return
	}
}

func (s Server) render500(w http.ResponseWriter, err error) {
	log.Printf("[ERROR] %v", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
