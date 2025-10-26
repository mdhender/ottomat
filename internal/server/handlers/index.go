// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/server/middleware"
)

// Index is a poor name for this handler, but Go sends routes that don't match
// anything to the handler for "/". That means that we have to do two things:
//  1. If the route really is "/", then redirect to either login or a dashboard.
//  2. Attempt to serve an asset.
func Index(assets fs.FS, client *ent.Client) http.HandlerFunc {
	startedAt := time.Now().UTC()
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: index: entered\n", r.Method, r.URL.Path)
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		asset := path.Clean(r.URL.Path)
		log.Printf("%s %s: index: asset %q\n", r.Method, r.URL.Path, asset)

		// root: route choice
		if asset == "/" {
			u, ok := middleware.GetUser(r.Context())
			if !ok { // no session, so redirect to login
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			} else if u.Role == user.RoleAdmin { // redirect to admin dashboard
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}
			// redirect to user dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// --- static asset branch ---

		// normalize and guard the path.
		asset = strings.TrimPrefix(asset, "/")
		if asset == "." || strings.HasPrefix(asset, "..") {
			http.NotFound(w, r)
			return
		}

		info, err := fs.Stat(assets, asset)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				log.Printf("%s %s: index: not found: %v\n", r.Method, r.URL.Path, err)
				http.NotFound(w, r)
				return
			}
			log.Printf("%s %s: index: stat: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if info.IsDir() { // never serve directories
			log.Printf("%s %s: index: directory\n", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		} else if !info.Mode().IsRegular() { // never serve special files
			log.Printf("%s %s: index: special file\n", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// determine mod time; adjust for some embedded file systems
		modTime := info.ModTime()
		if modTime.IsZero() {
			modTime = startedAt
		}

		etag := weakETag(info)
		immutable := isFingerprinted(asset) // e.g. "app.abc123.js"
		setCachingHeaders(w, etag, modTime, immutable)

		// Short-circuit on If-None-Match
		if inm := r.Header.Get("If-None-Match"); inm != "" && etagMatch(inm, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Optionally support If-Modified-Since too (cooperates with ServeContent)
		if ims := r.Header.Get("If-Modified-Since"); ims != "" && !modTime.IsZero() {
			if t, err := time.Parse(http.TimeFormat, ims); err == nil && !modTime.After(t) {
				// not modified
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		// open the asset
		fp, err := assets.Open(asset)
		if err != nil {
			log.Printf("%s %s: index: open error: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer fp.Close()

		// best-effort content-type by extension first
		setContentType := false
		if ct := mime.TypeByExtension(path.Ext(asset)); ct != "" {
			w.Header().Set("Content-Type", ct)
			setContentType = true
		}

		// fast path: already a ReadSeeker
		if rs, ok := fp.(io.ReadSeeker); ok {
			http.ServeContent(w, r, asset, modTime, rs)
			return
		}

		// fallback: read into memory for a ReadSeeker
		// (Use size hint if available to avoid extra allocs)
		var data []byte
		if size := info.Size(); size > 0 && size <= 8<<20 { // smallish files: allocate once
			data = make([]byte, 0, size)
			buf := bytes.NewBuffer(data)
			if _, err := io.Copy(buf, fp); err != nil {
				log.Printf("%s %s: index: read error: %v", r.Method, r.URL.Path, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			data = buf.Bytes()
		} else {
			// size unknown/large: still read (you could stream via FileServer instead)
			var err error
			data, err = io.ReadAll(fp)
			if err != nil {
				log.Printf("%s %s: index: readAll error: %v", r.Method, r.URL.Path, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// If we didn't set a content-type, detect from bytes
		if !setContentType && len(data) > 0 {
			w.Header().Set("Content-Type", http.DetectContentType(data))
		}

		http.ServeContent(w, r, asset, modTime, bytes.NewReader(data))
	}
}

func etagMatch(header, current string) bool {
	for _, tok := range strings.Split(header, ",") {
		if strings.TrimSpace(tok) == current || strings.TrimSpace(tok) == "*" {
			return true
		}
	}
	return false
}

func weakETag(info fs.FileInfo) string {
	// RFC 7232 allows opaque tokens; W/ marks it as weak.
	return fmt.Sprintf(`W/"%x-%x"`, info.ModTime().Unix(), info.Size())
}

func setCachingHeaders(w http.ResponseWriter, etag string, modTime time.Time, immutable bool) {
	w.Header().Set("ETag", etag)
	if immutable {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else { // tweak to taste
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	if !modTime.IsZero() {
		w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	}
}

func isFingerprinted(name string) bool {
	// todo: implement when we have a build process for assets
	return false
}
