package tpl

import (
	"html/template"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
)

type NonCachingRenderer struct {
	FS    fs.FS // filesystem containing the views
	Funcs template.FuncMap
}

func NewNonCaching(fsys fs.FS, funcs template.FuncMap) *NonCachingRenderer {
	nc := &NonCachingRenderer{FS: fsys, Funcs: funcs}
	return nc
}

func (nc *NonCachingRenderer) Glob() []string {
	layouts, partials, pages, fragments, err := findGoHtmlTemplates(nc.FS)
	if err != nil {
		return nil
	}
	spread := func(ss ...[]string) []string {
		var list []string
		for _, s := range ss {
			list = append(list, s...)
		}
		return list
	}
	return spread(layouts, partials, pages, fragments)
}

func (r *NonCachingRenderer) Page(w io.Writer, entry string, data any) error {
	return r.render(w, "pages/"+entry, data)
}

func (r *NonCachingRenderer) Frag(w io.Writer, entry string, data any) error {
	return r.render(w, "frags/"+entry, data)
}

func (r *NonCachingRenderer) render(w io.Writer, entry string, data any) error {
	t, err := parseAll(r.FS, r.Funcs, entry, r.Glob()...)
	if err != nil {
		log.Printf("render: nc: entry %q: parseAll %v\n", entry, err)
		return err
	}
	return t.ExecuteTemplate(w, entry, data)
}

// Preload is a no-op in development
func (r *NonCachingRenderer) Preload() error {
	return nil
}

// findGoHtmlTemplates recursively walks an fs.FS and returns all .gohtml files.
func findGoHtmlTemplates(fsys fs.FS) (layouts, partials, pages, fragments []string, err error) {
	find := func(fsys fs.FS, root string) (matches []string, err error) {
		err = fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil { // Propagate errors (permissions, etc.)
				return err
			} else if d.IsDir() {
				return nil
			} else if ok, _ := filepath.Match("*.gohtml", filepath.Base(path)); ok {
				matches = append(matches, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		slices.Sort(matches)
		return matches, nil
	}

	if layouts, err = find(fsys, "layouts"); err != nil {
		return nil, nil, nil, nil, err
	} else if partials, err = find(fsys, "partials"); err != nil {
		return nil, nil, nil, nil, err
	} else if pages, err = find(fsys, "pages"); err != nil {
		return nil, nil, nil, nil, err
	} else if fragments, err = find(fsys, "frags"); err != nil {
		return nil, nil, nil, nil, err
	}
	return layouts, partials, pages, fragments, nil
}
