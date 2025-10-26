package handlers

import "github.com/mdhender/ottomat/internal/server/templates"

func NewTemplateLoader(devMode bool) *templates.Loader {
	return templates.NewLoader(devMode)
}
