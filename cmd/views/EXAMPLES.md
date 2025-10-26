# cmd/views — Example App

This example demonstrates a minimal **HTMX + Go** setup with:

1. **Split filesystems** for `public/` and `views/` provided by the **root `ottomat` package**:
   - `GetPublicFS(FSConfig)` returns an `fs.FS` rooted at `public/`.
   - `GetViewsFS(FSConfig)` returns an `fs.FS` rooted at `views/`.
   - Both support **Live** (disk) and **Embedded** (embed.FS) modes.

2. A pluggable **template loader** in `internal/tpl` with a clean interface:
   - `NonCachingLoader` (development): reparses templates on every request for fast feedback.
   - `CachingLoader` (production): caches parsed templates; parses *all* entries and returns all errors (fail fast).

3. **HTMX fragment vs full page** responses:
   - `/users` renders a full page on normal requests and returns only a `<tbody>` row fragment on HTMX requests.
   - The fragment lives under `views/frags/users/table_rows.gohtml`.

---

## How to run

The example programs runs two servers, one caching views and the other refreshing views on each request.
Assets are served from the public/ tree.

```bash
go run ./cmd/views
# http://localhost:8080/       → non-caching, serves from public/
# http://localhost:8080/users  → non-caching, full page; HTMX swaps on search
# http://localhost:8181/       → caching, serves from public/
# http://localhost:8181/users  → caching, full page; HTMX swaps on search
# Static files are served from the embedded public/ tree.
```

---

## File map (key parts)

```
.
├── fs.go                              # root ottomat package (GetPublicFS/GetViewsFS)
├── public/                            # static assets (served at /)
├── views/
│   ├── layouts/base.gohtml            # layout with {{block "content" .}}
│   ├── partials/flash.gohtml
│   ├── pages/users/index.gohtml       # full page (uses base)
│   └── frags/users/table_rows.gohtml  # fragment for HTMX swaps
└── internal/views/
    └── loaders.go                     # implementation
```

---

## Notes

- The loader is intentionally **agnostic** of where files live; it only sees an `fs.FS`. That keeps tests easy and lets you swap backends without changing handlers.
- The example uses `.gohtml` extensions and short globs because the `views` FS is already rooted at `views/`.
- If you already have existing static pages under `public/`, you can convert them gradually by adding `.gohtml` versions under `views/` and routing to them.
