// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"
)

// Loader is a template loader.
//
// Example
//
//	if t, err := l.Load(name); err != nil {
//		return err
//	} else if err = t.ExecuteTemplate(w, name, data); err != nil {
//		return err
//	}
//
// Loaders assume that the name of each template is the same as the file name
// minus the ".gohtml" extension. Odd things may happen if multiple templates
// define the same name.
type Loader interface {
	// Load returns a template.
	Load(name string) (*template.Template, error)

	// Execute is a helper to load and execute a template.
	Execute(name string, data any) (*bytes.Buffer, error)
}

// CachingLoader can be created with either an embedded or "live" file system.
// Either way, it will scan the file system for the list of Go template files to load.
type CachingLoader struct {
	// the name of the template is the key into the cache
	cache map[string]*cachedTemplate
}

type cachedTemplate struct {
	t   *template.Template
	err error
}

// NewCachingLoader returns a CachingLoader using the Go template files
// in the file system. It preloads all the pages/ and frags/ views, returning
// all the errors found.
//
// We assume that the name of all templates is the same as the file name
// minus the ".gohtml" extension. Odd things will happen if multiple templates
// define the same name.
//
// Note: the preloader caches only pages/ and frags/, since those are the only
// views that we expect you to load directly.
func NewCachingLoader(fsys fs.FS, funcs template.FuncMap) (Loader, []error) {
	// find all the Go template files on the file system
	layouts, partials, pages, fragments, err := findGoHtmlTemplates(fsys)
	if err != nil {
		return nil, []error{err}
	}
	var files []string
	for _, list := range [][]string{layouts, partials, pages, fragments} {
		files = append(files, list...)
	}

	l := &CachingLoader{
		cache: make(map[string]*cachedTemplate),
	}

	// load all the pages/ and frags/ into the cache
	var errs []error
	for _, fileName := range files {
		matches := strings.HasPrefix(fileName, "pages/") || strings.HasPrefix(fileName, "frags/")
		if !matches {
			continue
		}
		name := strings.TrimSuffix(fileName, ".gohtml")
		t, err := parseAll(fsys, funcs, name, files...)
		if err != nil {
			log.Printf("views: preload %q: parse %q: %v", fileName, name, err)
			errs = append(errs, err)
		}
		l.cache[name] = &cachedTemplate{t: t, err: err}
	}

	if len(errs) != 0 {
		return l, errs
	}
	return l, nil
}

// Load returns a template and any parsing errors from the cache, using the name as the key.
func (l *CachingLoader) Load(name string) (*template.Template, error) {
	ct, ok := l.cache[name]
	if !ok {
		return nil, fmt.Errorf("%s: not found", name)
	}
	return ct.t, ct.err
}

// Execute uses Load() to fetch a template and execute it, returning a buffer or an error.
func (l *CachingLoader) Execute(name string, data any) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if t, err := l.Load(name); err != nil {
		return nil, err
	} else if err = t.ExecuteTemplate(buf, name, data); err != nil {
		return nil, err
	}
	return buf, nil
}

// NonCachingLoader should be created with a "live" file system.
type NonCachingLoader struct {
	fsys  fs.FS
	funcs template.FuncMap
}

// NewNonCachingLoader returns a NonCachingLoader using the Go template files
// in the file system.
func NewNonCachingLoader(fsys fs.FS, funcs template.FuncMap) (Loader, []error) {
	// find all the Go template files on the file system
	layouts, partials, pages, fragments, err := findGoHtmlTemplates(fsys)
	if err != nil {
		return nil, []error{err}
	}
	var files []string
	for _, list := range [][]string{layouts, partials, pages, fragments} {
		files = append(files, list...)
	}

	l := &NonCachingLoader{
		fsys:  fsys,
		funcs: funcs,
	}
	return l, nil
}

// Load loads a template from the file system and parses it, returning any errors.
func (l *NonCachingLoader) Load(name string) (*template.Template, error) {
	files, err := findTemplates(l.fsys)
	if err != nil {
		return nil, err
	}
	return template.New(name).Funcs(l.funcs).ParseFS(l.fsys, files...)
}

// Execute uses Load() to fetch a template and execute it, returning a buffer or an error.
func (l *NonCachingLoader) Execute(name string, data any) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if t, err := l.Load(name); err != nil {
		return nil, err
	} else if err = t.ExecuteTemplate(buf, name, data); err != nil {
		return nil, err
	}
	return buf, nil
}

// findTemplates recursively walks a filesystem looking for ".gohtml" files.
// It returns a sorted slice containing all the files it found.
func findTemplates(fsys fs.FS) (files []string, err error) {
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil { // Propagate errors (permissions, etc.)
			return err
		} else if d.IsDir() {
			return nil
		} else if ok, _ := filepath.Match("*.gohtml", filepath.Base(path)); ok {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	return files, nil
}
