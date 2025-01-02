package pages

import (
	"edu-portal/app/server/middleware"
	"edu-portal/app/server/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/AlekSi/pointer"
)

func (p Pages) Users(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.UserFromCtx(r.Context())
	if !exists {
		utils.Render500(w, fmt.Errorf("user not found"))
		return
	}
	users, err := p.Store.GetUsers(r.Context())
	if err != nil {
		utils.Render500(w, fmt.Errorf("user not found"))
		return
	}

	formatted := make([]struct {
		ID        int
		Telegram  string
		Status    string
		Role      string
		CreatedAt string
	}, len(users))
	for idx, u := range users {
		formatted[idx] = struct {
			ID        int
			Telegram  string
			Status    string
			Role      string
			CreatedAt string
		}{
			ID:       u.Id,
			Telegram: pointer.Get(u.TgUsername),
			Status:   string(u.Status),
			Role: func() string {
				if u.IsStaff {
					return "staff"
				} else {
					return "user"
				}
			}(),
			CreatedAt: u.CreatedAt.Format(time.DateTime),
		}
	}

	p.render(w, 200, "users.tmpl.html", "", map[string]interface{}{
		"user":  user,
		"users": formatted,
	})
}
