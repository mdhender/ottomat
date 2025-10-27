package server

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/internal/server/handlers"
	"github.com/mdhender/ottomat/internal/server/middleware"
	"github.com/mdhender/ottomat/internal/views"
)

type Server struct {
	http.Server
	viewLoader views.Loader
}

func New(addr string, client *ent.Client, devMode bool, avoidAutofill, visiblePasswords bool, assetsFS, viewsFS fs.FS) *Server {
	s := &Server{}
	s.Addr = addr

	var errs []error
	if devMode {
		s.viewLoader, errs = views.NewNonCachingLoader(viewsFS, nil)
	} else {
		s.viewLoader, errs = views.NewCachingLoader(viewsFS, nil)
	}
	if errs != nil {
		for _, err := range errs {
			log.Printf("server: views: load %v\n", err)
		}
		log.Fatalf("server: views: load failed\n")
	}
	sessionMW := middleware.Session(client)
	authMW := middleware.Auth()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /login", handlers.LoginPage(s.viewLoader, avoidAutofill, visiblePasswords))
	mux.HandleFunc("POST /login", handlers.PostLogin(client))
	mux.HandleFunc("POST /logout", handlers.PostLogout(client))

	mux.Handle("GET /admin", sessionMW(authMW(handlers.AdminDashboard(client, s.viewLoader))))
	mux.Handle("POST /admin/users", sessionMW(authMW(handlers.CreateUser(client))))
	mux.Handle("DELETE /admin/users/{id}", sessionMW(authMW(handlers.DeleteUser(client))))
	mux.Handle("GET /dashboard", sessionMW(authMW(http.HandlerFunc(handlers.Dashboard))))

	// home page and assets. per the Go blog, "As a special case, GET also matches HEAD."
	mux.Handle("GET /", sessionMW(handlers.Index(assetsFS, client)))

	s.Handler = mux

	// Wrap with logging middleware if in development mode
	if devMode {
		s.Handler = middleware.Logging()(s.Handler)
	}

	return s
}
