# Amp Agent: OttoMat

We are going to create the OttoMap web server.

1. Go web server
2. HTMX + AlpineJS + TailwindCSS
3. github.com/maloquacious/semver for semantic versioning
4. Ent + Altlas for ORM
5. modernc/sqlite for data store
6. Use the Go standard library as much as possible

## Commands
* CLI command:
  * Build CLI: `go build -o dist/local/ottomat .`
  * Version info: `dist/local/ottomat version`
  * Tests: `go test ./...`
  * Format code: `go fmt ./...`
  * Build for Linux: get version then `GOOS=linux GOARCH=amd64 go build -o dist/linux/ottomat-${VERSION} .`

## Code Style
- Standard Go formatting using `gofmt`
- Imports organized by stdlib first, then external packages
- Error handling: return errors to caller, log.Fatal only in main
- Function comments use Go standard format `// FunctionName does X`
- Variable naming follows camelCase
- File structure follows standard Go package conventions
- Type naming follows standard Go conventions (no special suffixes)
