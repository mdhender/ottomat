package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"sync"
)

//go:embed web/templates
var templatesFS embed.FS

type Loader interface {
}

type Template struct {
	name string
}

func (t *Template) Name() string {
	return t.name
}

func (t *Template) Execute(wr io.Writer, data any) error {
	return fmt.Errorf("%s: not implemented", t.Name())
}

type Loader struct {
	devMode   bool
	cached    *template.Template
	cacheLock sync.RWMutex
}

func NewLoader(devMode bool) *Loader {
	return &Loader{
		devMode: devMode,
	}
}

func (l *Loader) Load() (*template.Template, error) {
	if l.devMode {
		// In dev mode, always reload templates
		return l.loadTemplates()
	}

	// In production, use cached templates
	l.cacheLock.RLock()
	if l.cached != nil {
		cached := l.cached
		l.cacheLock.RUnlock()
		return cached, nil
	}
	l.cacheLock.RUnlock()

	// Cache miss, load and cache
	l.cacheLock.Lock()
	defer l.cacheLock.Unlock()

	// Double-check after acquiring write lock
	if l.cached != nil {
		return l.cached, nil
	}

	tmpl, err := l.loadTemplates()
	if err != nil {
		return nil, err
	}

	l.cached = tmpl
	return tmpl, nil
}

func (l *Loader) loadTemplates() (*template.Template, error) {
	return template.ParseFS(templatesFS, "web/templates/*.html")
}
