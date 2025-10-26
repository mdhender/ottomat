// Package main implements examples for caching and non-caching HTMX servers.
package main

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mdhender/ottomat"
	tpl "github.com/mdhender/ottomat/internal/views"
)

type User struct {
	Name  string
	Email string
}

type PageData struct {
	Q       string
	Users   []User
	Version string
}

func main() {
	log.SetFlags(log.Lshortfile)

	//log.Println("[views] dev mode: NonCaching HTMX server (reparse each request)")
	nonCachingPublicFS := ottomat.GetPublicFS(ottomat.FSConfig{Mode: ottomat.Live})
	nonCachingViewsFS := ottomat.GetViewsFS(ottomat.FSConfig{Mode: ottomat.Live})
	nonCachingRenderer := tpl.NewNonCaching(nonCachingViewsFS, funcs())
	err := nonCachingRenderer.Preload()
	if err != nil {
		log.Fatalf("nonCachingServer: %v\n", err)
	}
	nonCachingLoader, errs := tpl.NewNonCachingLoader(nonCachingViewsFS, funcs())
	if errs != nil {
		for _, err := range errs {
			log.Printf("nonCachingLoader: %v\n", err)
		}
		log.Fatalf("nonCachingLoader: failed\n")
	} else if buf, err := nonCachingLoader.Execute("pages/users/index", nil); err != nil {
		log.Fatalf("nonCachingLoader: %v\n", err)
	} else {
		log.Printf("nonCachingLoader: %q\n", buf.Bytes())
	}
	nonCachingServer, err := newServerWithLoader(":8080", nonCachingPublicFS, nonCachingLoader)
	if err != nil {
		log.Fatalf("nonCachingServer: %v\n", err)
	}

	//log.Println("[views] prod mode: Caching HTMX server (cache views)")
	cachingPublicFS := ottomat.GetPublicFS(ottomat.FSConfig{Mode: ottomat.Embedded})
	cachingViewsFS := ottomat.GetViewsFS(ottomat.FSConfig{Mode: ottomat.Embedded})
	cachingLoader, errs := tpl.NewCachingLoader(cachingViewsFS, funcs())
	if errs != nil {
		for _, err := range errs {
			log.Printf("cachingLoader: %v\n", err)
		}
		log.Fatalf("cachingLoader: failed\n")
	}
	cachingServer, err := newServerWithLoader(":8181", cachingPublicFS, cachingLoader)
	if err != nil {
		log.Fatalf("cachingServer: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		log.Printf("cachingServer listening on %s …", cachingServer.Addr)
		_ = cachingServer.ListenAndServe()
	}()

	go func() {
		log.Printf("nonCachingServer listening on %s …", nonCachingServer.Addr)
		_ = nonCachingServer.ListenAndServe()
	}()

	defer func() {
		if err := cachingServer.Shutdown(ctx); err != nil {
			fmt.Println("cachingServer shutting down: error: ", err)
		} else {
			fmt.Println("cachingServer shut down")
		}
		if err := nonCachingServer.Shutdown(ctx); err != nil {
			fmt.Println("nonCachingServer shutting down: error: ", err)
		} else {
			fmt.Println("nonCachingServer shut down")
		}
		fmt.Println("service has shutdown")
	}()

	sig := <-sigs
	fmt.Println(sig)
	fmt.Println("shutting down")

	cancel()
}

type server struct {
	http.Server
}

func newServer(addr string, pubFS fs.FS, render tpl.Renderer) (*server, error) {
	s := &server{}
	s.Addr = addr

	mux := http.NewServeMux()

	// static server
	mux.Handle("GET /", http.FileServer(http.FS(pubFS)))

	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		data := PageData{Q: q, Users: filterUsers(sampleUsers(), q), Version: ottomat.Version().String()}

		renderer, entry := render.Page, "users/index"
		if r.Header.Get("HX-Request") == "true" {
			renderer, entry = render.Frag, "users/table_rows"
		}
		if err := renderer(w, entry, data); err != nil {
			handleErr(w, err)
		}
	})

	s.Handler = mux

	return s, nil
}

func newServerWithLoader(addr string, pubFS fs.FS, loader tpl.Loader) (*server, error) {
	s := &server{}
	s.Addr = addr

	mux := http.NewServeMux()

	// static server
	mux.Handle("GET /", http.FileServer(http.FS(pubFS)))

	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		data := PageData{Q: q, Users: filterUsers(sampleUsers(), q), Version: ottomat.Version().String()}

		var name string
		if r.Header.Get("HX-Request") == "true" {
			name = "frags/users/table_rows"
		} else {
			name = "pages/users/index"
		}
		buf, err := loader.Execute(name, data)
		if err != nil {
			log.Printf("%s %s: %s: %v\n", r.Method, r.URL.Path, name, err)
			http.Error(w, "<p>TEMPLATE ERROR</p>", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	})

	s.Handler = mux

	return s, nil
}

func funcs() template.FuncMap {
	return template.FuncMap{}
}

func handleErr(w http.ResponseWriter, err error) {
	http.Error(w, "TEMPLATE ERROR:\n"+err.Error(), http.StatusInternalServerError)
}

func sampleUsers() []User {
	return []User{
		{"Ada Lovelace", "ada@examples.dev"},
		{"Alan Turing", "alan@examples.dev"},
		{"Grace Hopper", "grace@examples.dev"},
		{"Donald Knuth", "donald@examples.dev"},
	}
}

func filterUsers(users []User, q string) []User {
	if q == "" {
		return users
	}
	q = strings.ToLower(q)
	out := users[:0]
	for _, u := range users {
		if strings.Contains(strings.ToLower(u.Name), q) || strings.Contains(strings.ToLower(u.Email), q) {
			out = append(out, u)
		}
	}
	return out
}
