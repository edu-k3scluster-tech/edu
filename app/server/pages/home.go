package pages

import (
	"edu-portal/app/server/middleware"
	"edu-portal/app/server/utils"
	"fmt"
	"net/http"
)

func (p Pages) Home(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.UserFromCtx(r.Context())
	if !exists {
		utils.Render500(w, fmt.Errorf("user not found"))
		return
	}
	p.render(w, 200, "home.tmpl.html", "", map[string]interface{}{
		"user": user,
	})
}
