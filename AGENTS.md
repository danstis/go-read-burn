# AGENTS.md - AI Agent Guidelines for go-read-burn

> A Go-based "read once and burn" secret sharing web application, inspired by Read2Burn.

## Project Overview

- **Language**: Go 1.20+
- **Type**: Web application (HTTP server with BoltDB storage)
- **Architecture**: Standard Go project layout (`cmd/`, `internal/`)
- **Dependencies**: gorilla/mux (routing), boltdb/bolt (storage), kelseyhightower/envconfig (configuration)

## Build, Test, and Lint Commands

### Building

```bash
# Build all packages
go build -v ./...

# Run locally with version info (requires gitversion)
make run

# Build with ldflags for version injection
go build -ldflags "-s -w -X 'main.version=VERSION' -X 'main.commit=COMMIT' -X 'main.date=DATE'" ./cmd/go-read-burn
```

### Testing

```bash
# Run all tests
go test ./...

# Run all tests with verbose output
go test -v ./...

# Run tests with race detection and coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run a single test by name
go test -v -run TestCreateDBDir ./cmd/go-read-burn

# Run tests in a specific package
go test -v ./cmd/go-read-burn/...

# Run tests with JSON output (for CI)
go test -v -race -coverprofile=coverage.out -covermode=atomic -json ./... | tee test-report.out
```

### Linting

```bash
# Run go vet
go vet ./...

# Run golangci-lint (used in CI)
golangci-lint run

# Run golangci-lint with checkstyle output
golangci-lint run --out-format checkstyle:golint-report.out
```

### Docker

```bash
# Build and run with Docker Compose
make up
# or
docker compose --project-directory deploy up --build --remove-orphans
```

## Code Style Guidelines

### Import Organization

Imports should be organized in groups separated by blank lines:

1. Standard library imports
2. Third-party imports

```go
import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/boltdb/bolt"
    "github.com/gorilla/mux"
    "github.com/kelseyhightower/envconfig"
)
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Packages | lowercase, short | `version`, `main` |
| Exported functions | PascalCase, descriptive | `IndexHandler`, `CreateHandler` |
| Unexported functions | camelCase | `loadConfig`, `openDB`, `createDBDir` |
| Variables | camelCase | `dbPath`, `listenPort` |
| Constants | camelCase for unexported | `version`, `commit`, `date` |
| Struct types | PascalCase | `Config` |
| Struct fields | PascalCase (exported) | `DBPath`, `ListenPort` |

### Error Handling

- Use `fmt.Errorf` with `%w` for error wrapping
- Use `log.Fatalf` for unrecoverable startup errors
- Return errors to callers; let callers decide how to handle them
- Check errors immediately after function calls

```go
// Wrapping errors with context
func openDB(dbPath string) (*bolt.DB, error) {
    err := createDBDir(dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create database directory: %w", err)
    }
    // ...
}

// Fatal errors at startup
if err != nil {
    log.Fatalf("failed to open DB: %v", err)
}
```

### Function Design

- Keep functions focused and small
- Extract logical units into separate functions
- Use descriptive function names that indicate purpose

```go
func main() {
    config, err := loadConfig()
    db, err = openDB(config.DBPath)
    r := mux.NewRouter()
    setupRoutes(r)
    templates, err = parseTemplates()
    srv := createServer(config.ListenHost, config.ListenPort, r)
    startServer(srv)
    shutdownServer(srv, db)
}
```

### Configuration

- Use environment variables via `envconfig`
- Prefix environment variables with `GRB_`
- Use struct tags for defaults and naming

```go
type Config struct {
    DBPath     string `default:"db/secrets.db" split_words:"true"`
    ListenPort string `default:"80" split_words:"true"`
    ListenHost string `default:"0.0.0.0" split_words:"true"`
}
```

### HTTP Handlers

- Use `http.HandlerFunc` signature
- Handle errors within handlers, return appropriate HTTP status
- Use `templates.ExecuteTemplate` for HTML responses

```go
func IndexHandler(w http.ResponseWriter, r *http.Request) {
    if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
        http.Error(w, "error generating json: "+err.Error(), 500)
        return
    }
}
```

### Testing Patterns

- Use table-driven tests where appropriate
- Use `httptest` for HTTP handler testing
- Create temporary directories for file-based tests
- Clean up resources with `defer`

```go
func TestCreateDBDir(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "test")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    err = createDBDir(tempDir + "/db/secrets.db")
    if err != nil {
        t.Errorf("Failed to create directory: %v", err)
    }
}
```

## Project Structure

```
go-read-burn/
├── cmd/
│   └── go-read-burn/       # Main application
│       ├── main.go         # Entry point
│       ├── main_test.go    # Tests
│       ├── views/          # HTML templates (embedded)
│       ├── static/         # Static assets (embedded)
│       └── dockerfile
├── internal/               # Private packages
│   └── version/
├── deploy/                 # Deployment configs
├── db/                     # Database storage
├── .github/workflows/      # CI/CD
└── go.mod
```

## Commit Message Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

| Prefix | Version Bump | Example |
|--------|--------------|---------|
| `feat:`, `feature:`, `minor:` | Minor | `feat: add password protection` |
| `fix:`, `patch:`, `hotfix:` | Patch | `fix: handle empty secrets` |
| `BREAKING CHANGES:`, `major:` | Major | `BREAKING CHANGES: new API format` |
| `build:`, `chore:`, `ci:`, `docs:`, `test:`, `refactor:` | None | `chore: update dependencies` |

## CI/CD Pipeline

- **Build**: `go build -v ./...`
- **Test**: `go test` with race detection and coverage
- **Lint**: golangci-lint
- **Quality**: SonarCloud analysis
- **Release**: GoReleaser (triggered on main branch)

## Editor Configuration

- Indent: 4 spaces (2 for YAML/JS)
- Charset: UTF-8
- Line endings: LF
- Trim trailing whitespace: Yes
- Insert final newline: Yes

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GRB_DB_PATH` | `db/secrets.db` | Path to BoltDB database |
| `GRB_LISTEN_PORT` | `80` | HTTP server port |
| `GRB_LISTEN_HOST` | `0.0.0.0` | HTTP server host |

## Key Dependencies

- **github.com/gorilla/mux**: HTTP router
- **github.com/boltdb/bolt**: Embedded key/value database
- **github.com/kelseyhightower/envconfig**: Environment configuration
