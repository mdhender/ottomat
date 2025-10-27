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

const (
	sessionCookieName = "ottomat_session"
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
			log.Printf("%s %s: fragment true\n", r.Method, r.URL.Path)
			http.Error(w, fmt.Sprintf("%s %s: view error: fragment not implemented", r.Method, r.URL.Path), http.StatusInternalServerError)
			return
		} else {
			log.Printf("%s %s: page true\n", r.Method, r.URL.Path)
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

func PostLogin(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		username := r.FormValue("username")
		password := r.FormValue("password")
		log.Printf("%s %s: username %q: password %q\n", r.Method, r.URL.Path, username, password)

		ctx := r.Context()
		u, err := client.User.
			Query().
			Where(user.Username(username)).
			Only(ctx)
		if err != nil {
			log.Printf("%s %s: query %v\n", r.Method, r.URL.Path, err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
			log.Printf("%s %s: bcrypt %v\n", r.Method, r.URL.Path, err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		token, err := auth.GenerateSessionToken()
		if err != nil {
			log.Printf("%s %s: genToken %v\n", r.Method, r.URL.Path, err)
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
			log.Printf("%s %s: setToken %v\n", r.Method, r.URL.Path, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		urlDashboard := "/dashboard" // assume user
		if u.Role == user.RoleAdmin {
			urlDashboard = "/admin"
		}
		if r.Header.Get("HX-Request") == "true" {
			// HTMX-specific header for full page redirect
			w.Header().Add("HX-Redirect", urlDashboard) // redirect to dashboard
			w.WriteHeader(http.StatusNoContent)         // no content to swap
			return
		}
		// traditional soft redirect
		http.Redirect(w, r, urlDashboard, http.StatusSeeOther)
	}
}

func PostLogout(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil {
			ctx := r.Context()
			client.Session.
				Delete().
				Where(session.Token(cookie.Value)).
				ExecX(ctx)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(-1 * time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		if r.Header.Get("HX-Request") == "true" {
			// HTMX-specific header for full page redirect
			w.Header().Add("HX-Redirect", "/login") // redirect to login page
			w.WriteHeader(http.StatusNoContent)     // no content to swap
			return
		}
		// traditional soft redirect
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
