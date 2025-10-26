package tpl

import (
	"html/template"
	"io"
	"io/fs"
	"log"
	"strings"
	"sync"
)

type CachingRenderer struct {
	cfg     Config
	mu      sync.RWMutex
	viewsFS fs.FS
	funcs   template.FuncMap
	files   []string
	cache   map[string]*template.Template
}

func NewCaching(cfg Config) *CachingRenderer {
	layouts, partials, pages, fragments, err := findGoHtmlTemplates(cfg.FS)
	if err != nil {
		return nil
	}
	var files []string
	for _, list := range [][]string{layouts, partials, pages, fragments} {
		files = append(files, list...)
	}
	return &CachingRenderer{
		cfg:     cfg,
		viewsFS: cfg.FS,
		files:   files,
		cache:   make(map[string]*template.Template),
	}
}

func (r *CachingRenderer) Page(w io.Writer, entry string, data any) error {
	return r.exec(w, "pages/"+entry, data)
}

func (r *CachingRenderer) Frag(w io.Writer, entry string, data any) error {
	return r.exec(w, "frags/"+entry, data)
}

func (r *CachingRenderer) Glob() []string {
	return r.files
}

func (r *CachingRenderer) exec(w io.Writer, entry string, data any) error {
	t, err := r.get(entry)
	if err != nil {
		return err
	}
	err = t.ExecuteTemplate(w, entry, data)
	if err != nil {
		return err
	}
	return nil
}

func (r *CachingRenderer) get(entry string) (*template.Template, error) {
	r.mu.RLock()
	if t, ok := r.cache[entry]; ok {
		r.mu.RUnlock()
		return t, nil
	}
	r.mu.RUnlock()

	t, err := parseAll(r.viewsFS, r.funcs, entry, r.Glob()...)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.cache[entry] = t
	r.mu.Unlock()
	return t, nil
}

// Preload parses & caches everything; logs all errors and returns the first one.
func (r *CachingRenderer) Preload() error {
	var firstErr error
	for _, fileName := range r.Glob() {
		matches := strings.HasPrefix(fileName, "pages/") || strings.HasPrefix(fileName, "frags/")
		if !matches {
			continue
		}
		entry := strings.TrimSuffix(fileName, ".gohtml")
		_, err := r.get(entry)
		if err != nil {
			log.Printf("cr: preload %q: parse %q: %v", fileName, entry, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
	}
	return firstErr
}
