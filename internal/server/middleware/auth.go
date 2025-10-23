package middleware

import (
	"net/http"

	"github.com/mdhender/ottomat/ent/user"
)

func Auth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := GetUser(r.Context())
			if !ok || u == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			if r.URL.Path == "/admin" || r.URL.Path == "/admin/users" {
				if u.Role != user.RoleAdmin {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
