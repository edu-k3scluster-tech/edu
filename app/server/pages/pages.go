package pages

import (
	"bytes"
	"context"
	"edu-portal/app"
	"edu-portal/app/server/utils"
	"fmt"
	"net/http"
	"text/template"
)

type Store interface {
	AssignOneTimeToken(ctx context.Context, token string) error
	GetLogs(ctx context.Context, userId int) ([]app.AuditLog, error)
	GetUsers(ctx context.Context) ([]app.User, error)
	GetUserCertificate(ctx context.Context, userId int) (*app.UserCertificate, error)
}

type Cluster interface {
	Config(ctx context.Context, certificate, privateKey []byte) (string, error)
}

type Pages struct {
	Templates map[string]*template.Template
	Store     Store
	Cluster   Cluster
}

func (p Pages) render(w http.ResponseWriter, status int, page, tmplName string, data any) {
	ts, ok := p.Templates[page]
	if !ok {
		utils.Render500(w, fmt.Errorf("the template %s does not exist", page))
		return
	}

	buf := new(bytes.Buffer)

	if tmplName == "" {
		tmplName = "base"
	}

	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		utils.Render500(w, err)
		return
	}

	w.WriteHeader(status)
	if _, err = buf.WriteTo(w); err != nil {
		utils.Render500(w, err)
		return
	}
}
