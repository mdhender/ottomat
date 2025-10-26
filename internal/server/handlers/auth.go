package handlers

import (
	"net/http"
	"time"

	"github.com/mdhender/ottomat"
	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/ent/session"
	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/auth"
	"github.com/mdhender/ottomat/internal/server/templates"
	"golang.org/x/crypto/bcrypt"
)

type LoginPageData struct {
	Title        string
	Version      string
	PasswordType string
}

func LoginPage(tmplLoader *templates.Loader, visiblePasswords bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		passwordType := "password"
		if visiblePasswords {
			passwordType = "text"
		}

		data := LoginPageData{
			Title:        "Login",
			Version:      ottomat.Version().String(),
			PasswordType: passwordType,
		}

		tmpl, err := tmplLoader.Load()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func Login(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		ctx := r.Context()
		u, err := client.User.
			Query().
			Where(user.Username(username)).
			Only(ctx)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		token, err := auth.GenerateSessionToken()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_, err = client.Session.
			Create().
			SetToken(token).
			SetExpiresAt(time.Now().Add(24 * time.Hour)).
			SetUser(u).
			Save(ctx)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func Logout(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil {
			ctx := r.Context()
			client.Session.
				Delete().
				Where(session.Token(cookie.Value)).
				ExecX(ctx)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(-1 * time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
