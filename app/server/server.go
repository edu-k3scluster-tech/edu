package server

import (
	"context"
	"edu-portal/app/cluster"
	"edu-portal/app/store"
	"edu-portal/app/usecases"
	"net/http"
	"text/template"

	"edu-portal/app/server/api"
	mdw "edu-portal/app/server/middleware"
	"edu-portal/app/server/pages"
	"edu-portal/app/server/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Server struct {
	templates     map[string]*template.Template
	authenticator *mdw.Authenticator
	store         *store.Store
	cluster       *cluster.Cluster
}

func New(secured bool, templates map[string]*template.Template, store *store.Store, cluster *cluster.Cluster) *Server {
	authenticator := &mdw.Authenticator{
		Resolver: store.GetUserByAuthToken,
	}

	return &Server{
		templates:     templates,
		authenticator: authenticator,
		store:         store,
		cluster:       cluster,
	}
}

func (s Server) Run(ctx context.Context) error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	pages := pages.Pages{Templates: s.templates, Store: s.store, Cluster: s.cluster}
	// HTML Users
	r.Group(func(r chi.Router) {
		r.Use(mdw.AuthMiddleware(mdw.AnyUser, utils.Redirect("/login"), utils.Redirect("/"), s.authenticator))

		r.Get("/", pages.Home)
		r.Get("/status", pages.Status)
		r.Get("/audit", pages.Audit)
	})

	// HTML Staff
	r.Group(func(r chi.Router) {
		r.Use(mdw.AuthMiddleware(mdw.OnlyStaff, utils.Redirect("/login"), utils.Redirect("/"), s.authenticator))

		r.Get("/users", pages.Users)
	})

	r.Get("/login", pages.Login)

	api := api.Api{Store: s.store, ActivateUC: usecases.NewActivateUser(s.store, s.cluster)}
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
			r.Use(mdw.AuthMiddleware(mdw.OnlyStaff, utils.Render401, utils.Render403, s.authenticator))

			r.Post("/users/{id}/activate", api.Activate)
		})
	})
	return http.ListenAndServe(":8000", r)
}
