package pages

import (
	"edu-portal/app/server/middleware"
	"fmt"
	"net/http"
)

func (p Pages) Status(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.UserFromCtx(r.Context())
	if !exists {
		p.render500(w, fmt.Errorf("user not found"))
		return
	}
	p.render(w, 200, "status.tmpl.html", "", map[string]interface{}{
		"user": user,
	})
}
