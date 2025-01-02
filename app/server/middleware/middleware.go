package middleware

import (
	"context"
	"edu-portal/app"
	"edu-portal/app/server/utils"
	"net/http"
)

type contextKey string

const userContextKey contextKey = "user"

func UserFromCtx(ctx context.Context) (app.User, bool) {
	user, ok := ctx.Value(userContextKey).(app.User)
	return user, ok
}

func UserToCtx(ctx context.Context, user app.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func AuthMiddleware(check func(*app.User) bool, on401, on403 func(w http.ResponseWriter, r *http.Request), authenticator *Authenticator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authenticator.IsAuthenticated(r.Context(), r)
			if err != nil {
				utils.Render500(w, err)
				return
			}
			if user == nil {
				on401(w, r)
				// http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if !check(user) {
				// http.Redirect(w, r, "/", http.StatusSeeOther)
				on403(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(UserToCtx(r.Context(), *user)))
		})
	}
}

func AnyUser(u *app.User) bool {
	return true
}

func OnlyStaff(u *app.User) bool {
	return u.IsStaff
}
