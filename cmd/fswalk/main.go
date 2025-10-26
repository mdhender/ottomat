package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// FindGoHTML recursively walks an fs.FS and returns all .gohtml files.
func FindGoHTML(fsys fs.FS) (layouts, partials, pages, fragments []string, err error) {
	if layouts, err = findTemplates(fsys, "layouts"); err != nil {
		return nil, nil, nil, nil, err
	} else if partials, err = findTemplates(fsys, "partials"); err != nil {
		return nil, nil, nil, nil, err
	} else if pages, err = findTemplates(fsys, "pages"); err != nil {
		return nil, nil, nil, nil, err
	} else if fragments, err = findTemplates(fsys, "frags"); err != nil {
		return nil, nil, nil, nil, err
	}
	return layouts, partials, pages, fragments, nil
}

func findTemplates(fsys fs.FS, root string) (matches []string, err error) {
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

func main() {
	// Example usage with the local filesystem
	layouts, partials, pages, fragments, err := FindGoHTML(os.DirFS("views"))
	if err != nil {
		panic(err)
	}
	for _, f := range layouts {
		fmt.Println(f)
	}
	for _, f := range partials {
		fmt.Println(f)
	}
	for _, f := range pages {
		fmt.Println(f)
	}
	for _, f := range fragments {
		fmt.Println(f)
	}
}
