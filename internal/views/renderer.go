// Package tpl provides a pluggable template renderer for Go html/template.
// It supports dev (non-caching, reparse) and prod (caching) with a Preload pass.
package views

import (
	"io"
)

type Renderer interface {
	// Page renders an entire html document. That is any request where HX-Request is not set to "true" in the header
	Page(w io.Writer, entry string, data any) error // entry under pages/

	// Frag renders a fragment - that is a request where HX-Request is set to "true" in the header
	Frag(w io.Writer, entry string, data any) error // entry under frags/

	Preload() error // dev: logs only; prod: returns first error
}
