package pages

import (
	"bytes"
	"context"
	"edu-portal/app"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

type Store interface {
	AssignOneTimeToken(ctx context.Context, token string) error
	GetLogs(ctx context.Context, userId int) ([]app.AuditLog, error)
}

type Pages struct {
	Templates map[string]*template.Template
	Store     Store
}

func (p Pages) render(w http.ResponseWriter, status int, page, tmplName string, data any) {
	ts, ok := p.Templates[page]
	if !ok {
		p.render500(w, fmt.Errorf("the template %s does not exist", page))
		return
	}

	buf := new(bytes.Buffer)

	if tmplName == "" {
		tmplName = "base"
	}

	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		p.render500(w, err)
		return
	}

	w.WriteHeader(status)
	if _, err = buf.WriteTo(w); err != nil {
		p.render500(w, err)
		return
	}
}

func (p Pages) render500(w http.ResponseWriter, err error) {
	log.Printf("[ERROR] %v", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}