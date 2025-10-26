package views

import (
	"html/template"
	"io/fs"
)

func def(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func parseAll(fsys fs.FS, funcs template.FuncMap, entry string, files ...string) (*template.Template, error) {
	//log.Printf("parseAll: entry %q\n", entry)
	fileName := entry + ".gohtml"
	//log.Printf("parseAll: entry %q: fileName %q\n", entry, fileName)
	t := template.New(entry).Funcs(funcs)
	// we put fileName last so that it overrides any accidental "define" mismatches
	templateFiles := append(files, fileName)
	return t.ParseFS(fsys, templateFiles...)
}
