package pages

import (
	"edu-portal/pkg/token"
	"fmt"
	"net/http"
)

func (p Pages) Login(w http.ResponseWriter, r *http.Request) {
	oneTimeToken := token.RandomToken()
	if err := p.Store.AssignOneTimeToken(r.Context(), oneTimeToken); err != nil {
		p.render500(w, err)
		return
	}

	p.render(w, 200, "login.tmpl.html", "", map[string]interface{}{
		"TgLink":       fmt.Sprintf("https://t.me/eduk3scluster_bot/?start=%s", oneTimeToken),
		"OneTimeToken": oneTimeToken,
	})
}
