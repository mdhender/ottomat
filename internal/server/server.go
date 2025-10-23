package server

import (
	"net/http"

	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/internal/server/handlers"
	"github.com/mdhender/ottomat/internal/server/middleware"
)

func New(client *ent.Client, devMode bool) http.Handler {
	mux := http.NewServeMux()

	sessionMW := middleware.Session(client)
	authMW := middleware.Auth()

	mux.HandleFunc("GET /login", handlers.LoginPage(devMode))
	mux.HandleFunc("POST /login", handlers.Login(client))
	mux.HandleFunc("POST /logout", handlers.Logout(client))

	mux.Handle("GET /", sessionMW(authMW(http.HandlerFunc(handlers.Dashboard))))
	mux.Handle("GET /admin", sessionMW(authMW(http.HandlerFunc(handlers.AdminDashboard(client)))))
	mux.Handle("POST /admin/users", sessionMW(authMW(http.HandlerFunc(handlers.CreateUser(client)))))
	mux.Handle("DELETE /admin/users/{id}", sessionMW(authMW(http.HandlerFunc(handlers.DeleteUser(client)))))

	// Wrap with logging middleware if in development mode
	var handler http.Handler = mux
	if devMode {
		handler = middleware.Logging()(handler)
	}

	return handler
}
