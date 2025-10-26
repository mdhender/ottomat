// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package middleware

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

// middleware to serve assets from an embedded or live file system.
// major differences are that the live file system checks for updates to
// files; the embedded file system does not.
//
// there's a hack here. this handler always forwards non-GET or root
// path requests. we are assuming that the application jumps through
// hoops to redirect / to a login or dashboard page and we're just
// here to serve assets.

func embeddedAssetsHandler(assets fs.FS, h http.Handler) http.HandlerFunc {
	// embedded file systems might not support timestamps, so fake one.
	fakeModTime := time.Now().UTC()

	type fileInfo struct {
		path    string
		modTime time.Time
		etag    string
	}
	// cachedFiles is a map of path to file info
	cachedFiles := map[string]fileInfo{}

	// cache all the file info
	err := fs.WalkDir(assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil { // Propagate errors (permissions, etc.)
			return err
		} else if d.IsDir() {
			return nil
		}
		mode, err := d.Info()
		if err != nil {
			return err
		} else if !mode.Mode().IsRegular() {
			return nil
		}
		modTime := mode.ModTime()
		if modTime.IsZero() {
			modTime = fakeModTime
		}
		// entry is a file, so compute ETag based on file content and cache it
		fp, err := assets.Open(path)
		if err != nil {
			return err
		}
		defer fp.Close()
		hasher := sha256.New()
		if _, err := io.Copy(hasher, fp); err != nil {
			return err
		}
		cachedFiles[path] = fileInfo{
			path:    path,
			modTime: modTime,
			etag:    fmt.Sprintf(`"%x"`, hasher.Sum(nil)),
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: middleware: embeddedAssets\n", r.Method, r.URL.Path)
		// janky - never serve the root path or any method other than GET
		if r.Method != http.MethodGet {
			h.ServeHTTP(w, r)
			return
		} else if r.URL.Path == "/" {
			log.Printf("%s %s: middleware: embeddedAssets: is the root\n", r.Method, r.URL.Path)
			h.ServeHTTP(w, r)
			return
		}

		asset := r.URL.Path
		log.Printf("%s %s: %q\n", r.Method, r.URL.Path, asset)

		fi, ok := cachedFiles[asset]
		if !ok {
			// not an asset
			log.Printf("%s %s: middleware: embeddedAssets: not an asset\n", r.Method, r.URL.Path)
			h.ServeHTTP(w, r)
			return
		}

		// let http handle the file
		fp, err := assets.Open(asset)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer fp.Close()

		// clients can use two headers to avoid downloading.
		// The first, If-Modified-Since, is a timestamp and is set by http.ServeContent
		// from the file's ModTime.
		//
		// The second, If-None-Match, is an ETag and is set by us. We compute the ETag
		// from a SHA256 of the file's content.
		//
		// If client sends an ETag header, compute the hash to see if it's still the same.
		if match := r.Header.Get("If-None-Match"); match != "" {
			if match == fi.etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		f, ok := fp.(io.ReadSeeker)
		if !ok {
			log.Printf("%s %s: %s: %v\n", r.Method, r.URL.Path, asset, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, r.URL.Path, fi.modTime, f)
	}
}

// isFileExists returns true if the path exists and is a regular file.
func isFileExists(assets fs.FS, path string) (bool, error) {
	sb, err := fs.Stat(assets, path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else if sb.IsDir() {
		return false, nil
	}
	return sb.Mode().IsRegular(), nil
}
