// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package views provides a pluggable template loader for Go html/template.
// It supports dev (non-caching, reparse) and prod (caching) with a Preload pass.
package views

import (
	"bytes"
	"html/template"
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
