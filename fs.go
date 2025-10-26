package ottomat

import (
	"embed"
	"io/fs"
	"os"
)

type FSMode uint8

const (
	Live     FSMode = iota // read from disk
	Embedded               // use embed.FS
)

type FSConfig struct {
	Mode       FSMode
	BaseDir    string // used in Live mode; default "."
	PublicRoot string // default "public"
	ViewsRoot  string // default "views"
}

// --- Embedded trees ---

//go:embed public/**
var publicEmbed embed.FS

//go:embed views/**
var viewsEmbed embed.FS

// GetPublicFS returns a filesystem rooted at the public/ directory.
func GetPublicFS(cfg FSConfig) fs.FS {
	if cfg.PublicRoot == "" {
		cfg.PublicRoot = "public"
	}
	if cfg.Mode == Live {
		base := cfg.BaseDir
		if base == "" {
			base = "."
		}
		return mustSub(os.DirFS(base), cfg.PublicRoot)
	}
	return mustSub(publicEmbed, cfg.PublicRoot)
}

// GetViewsFS returns a filesystem rooted at the views/ directory.
func GetViewsFS(cfg FSConfig) fs.FS {
	if cfg.ViewsRoot == "" {
		cfg.ViewsRoot = "views"
	}
	if cfg.Mode == Live {
		base := cfg.BaseDir
		if base == "" {
			base = "."
		}
		return mustSub(os.DirFS(base), cfg.ViewsRoot)
	}
	return mustSub(viewsEmbed, cfg.ViewsRoot)
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
