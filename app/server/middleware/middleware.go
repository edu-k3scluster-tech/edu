package middleware

import (
	"context"
	"edu-portal/app"
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

func AuthMiddleware(onerr func(w http.ResponseWriter, err error), check func(*app.User) bool, authenticator *Authenticator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authenticator.IsAuthenticated(r.Context(), r)
			if err != nil {
				onerr(w, err)
				return
			}
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if !check(user) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
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
