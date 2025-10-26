# OttoMat HTMX Template Demo (split FS)

This demo shows:
- Separate `fs.FS` for `public/` and `views/` via the root `ottomat` package.
- `internal/tpl` renderer with NonCaching (dev) and Caching (prod) + `Preload`.
- `.gohtml` templates with layouts/pages/frags/partials.

## Run (dev)
```bash
go run ./cmd/views --dev
# http://localhost:8181/  (public landing)
# http://localhost:8181/users  (HTMX demo)
```

## Run (prod-style, embedded)
```bash
go run ./cmd/views
# (Preload fails fast if a template is broken)
```

Adjust imports to your module path if needed.
