package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/ent/session"
)

const (
	sessionCookieName = "ottomat_session"
)

type contextKey string

const UserContextKey contextKey = "user"

func Session(client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			sess, err := client.Session.
				Query().
				Where(session.Token(cookie.Value)).
				Where(session.ExpiresAtGT(time.Now())).
				WithUser().
				Only(ctx)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user := sess.Edges.User
			if user != nil {
				ctx = context.WithValue(ctx, UserContextKey, user)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUser(ctx context.Context) (*ent.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*ent.User)
	return user, ok
}
