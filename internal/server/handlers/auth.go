package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mdhender/ottomat"
	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/ent/session"
	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/auth"
	"github.com/mdhender/ottomat/internal/views"
	"golang.org/x/crypto/bcrypt"
)

type LoginPageData struct {
	Title         string
	Version       string
	PasswordType  string
	AvoidAutofill bool
}

func LoginPage(view views.Loader, avoidAutofill, visiblePasswords bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		passwordType := "password"
		if visiblePasswords {
			passwordType = "text"
		}
		data := LoginPageData{
			Title:         "Login",
			Version:       ottomat.Version().String(),
			PasswordType:  passwordType,
			AvoidAutofill: avoidAutofill,
		}

		var name string
		if r.Header.Get("HX-Request") == "true" {
			// should not be supported?
			http.Error(w, fmt.Sprintf("<p>%s %s: view error: fragment not implemented</p>", r.Method, r.URL.Path), http.StatusInternalServerError)
			return
		} else {
			name = "pages/login"
		}
		buf, err := view.Execute(name, data)
		if err != nil {
			log.Printf("%s %s: %s: data %+v\n", r.Method, r.URL.Path, name, data)
			log.Printf("%s %s: %s: render %v\n", r.Method, r.URL.Path, name, err)
			http.Error(w, fmt.Sprintf("%s %s: %s: view error: %v", r.Method, r.URL.Path, name, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}
}

func Login(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
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
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
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
