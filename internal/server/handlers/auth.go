package handlers

import (
	"fmt"
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

func LoginPage(devMode bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		autocompleteAttrs := ""
		if devMode {
			autocompleteAttrs = `autocomplete="off" data-1p-ignore data-lpignore="true"`
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - OttoMat</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <script src="https://unpkg.com/alpinejs@3.14.8/dist/cdn.min.js" defer></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-900 text-white min-h-screen flex flex-col">
    <div class="flex-grow flex items-center justify-center">
        <div class="bg-gray-800 p-8 rounded-lg shadow-lg w-96">
            <h1 class="text-2xl font-bold mb-6 text-center">OttoMat Login</h1>
            <form hx-post="/login" hx-target="body" hx-swap="outerHTML" %s>
                <div class="mb-4">
                    <label for="username" class="block text-sm font-medium mb-2">Username</label>
                    <input type="text" id="username" name="username" required %s
                        class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:border-blue-500">
                </div>
                <div class="mb-6">
                    <label for="password" class="block text-sm font-medium mb-2">Password</label>
                    <input type="password" id="password" name="password" required autocomplete="current-password" %s
                        class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:border-blue-500">
                </div>
                <button type="submit" 
                    class="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 rounded transition">
                    Login
                </button>
            </form>
        </div>
    </div>
    %s
</body>
</html>`, autocompleteAttrs, autocompleteAttrs, autocompleteAttrs, templates.Footer(ottomat.Version().String()))
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
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
