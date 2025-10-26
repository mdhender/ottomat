package server

import (
	"io/fs"
	"net/http"

	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/internal/server/handlers"
	"github.com/mdhender/ottomat/internal/server/middleware"
	"github.com/mdhender/ottomat/internal/views"
)

type Server struct {
	http.Server
	viewRenderer views.Renderer
}

func New(addr string, client *ent.Client, devMode bool, visiblePasswords bool, assetsFS, viewsFS fs.FS) *Server {
	s := &Server{}
	s.Addr = addr

	if devMode {
		s.viewRenderer = views.NewNonCaching(viewsFS, nil)
	} else {
		s.viewRenderer = views.NewCaching(viewsFS, nil)
	}
	sessionMW := middleware.Session(client)
	authMW := middleware.Auth()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /login", handlers.LoginPage(s.viewRenderer, visiblePasswords))
	mux.HandleFunc("POST /login", handlers.Login(client))
	mux.HandleFunc("POST /logout", handlers.Logout(client))

	mux.Handle("GET /", sessionMW(authMW(http.HandlerFunc(handlers.Dashboard))))
	mux.Handle("GET /admin", sessionMW(authMW(http.HandlerFunc(handlers.AdminDashboard(client)))))
	mux.Handle("POST /admin/users", sessionMW(authMW(http.HandlerFunc(handlers.CreateUser(client)))))
	mux.Handle("DELETE /admin/users/{id}", sessionMW(authMW(http.HandlerFunc(handlers.DeleteUser(client)))))

	s.Handler = mux

	// Wrap with logging middleware if in development mode
	if devMode {
		s.Handler = middleware.Logging()(s.Handler)
	}

	return s
}
