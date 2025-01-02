package pages

import (
	"edu-portal/app/server/middleware"
	"edu-portal/app/server/utils"
	"fmt"
	"net/http"
	"time"
)

func (p Pages) Audit(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.UserFromCtx(r.Context())
	if !exists {
		utils.Render500(w, fmt.Errorf("user not found"))
		return
	}
	logs, err := p.Store.GetLogs(r.Context(), user.Id)
	if err != nil {
		utils.Render500(w, err)
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

	p.render(w, 200, "audit.tmpl.html", "", map[string]interface{}{
		"user": user,
		"logs": formatted,
	})
}
