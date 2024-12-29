package server

import (
	"context"
	"edu-portal/app/store"
	"log"
	"net/http"
	"text/template"

	"edu-portal/app/server/api"
	mdw "edu-portal/app/server/middleware"
	"edu-portal/app/server/pages"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Server struct {
	templates     map[string]*template.Template
	authenticator *mdw.Authenticator
	store         *store.Store
}

func New(secured bool, templates map[string]*template.Template, store *store.Store) *Server {
	return &Server{
		templates: templates,
		authenticator: &mdw.Authenticator{
			Secured:  secured,
			Resolver: store.GetUserByAuthToken,
		},
		store: store,
	}
}

func (s Server) Run(ctx context.Context) error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	pages := pages.Pages{Templates: s.templates, Store: s.store}
	// HTML Users
	r.Group(func(r chi.Router) {
		r.Use(mdw.AuthMiddleware(s.render500, mdw.AnyUser, s.authenticator))

		r.Get("/", pages.Home)
		r.Get("/status", pages.Status)
		r.Get("/audit", pages.Audit)
	})

	// HTML Staff
	r.Group(func(r chi.Router) {
		r.Use(mdw.AuthMiddleware(s.render500, mdw.OnlyStaff, s.authenticator))

		r.Get("/users", pages.Users)
	})

	r.Get("/login", pages.Login)

	api := api.Api{Store: s.store}
	// REST
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Use(render.SetContentType(render.ContentTypeJSON))

		// Public
		r.Group(func(r chi.Router) {
			r.Post("/auth", api.Auth)
		})

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(mdw.AuthMiddleware(s.render500, mdw.OnlyStaff, s.authenticator))

			r.Post("/users/{id}/activate", api.Activate)
		})
	})
	return http.ListenAndServe(":8000", r)
}

func (s Server) render500(w http.ResponseWriter, err error) {
	log.Printf("[ERROR] %v", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
