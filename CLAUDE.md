# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**go-read-burn** is a Go-based "read once and burn" secret sharing web application inspired by Read2Burn. It implements zero-knowledge encryption where secrets are encrypted client-side and the server never has access to decryption keys.

- **Language**: Go 1.24.0+
- **Type**: HTTP web server with embedded templates and static assets
- **Database**: BoltDB (key-value store)
- **Encryption**: AES-256-GCM with scrypt key derivation
- **Architecture**: Standard Go project layout (`cmd/`, `internal/`)

## Build, Test, and Run Commands

### Building

```bash
# Build all packages
go build -v ./...

# Run locally (basic, no version info)
go run ./cmd/go-read-burn

# Run with version metadata (requires gitversion)
make run

# Build with version injection
go build -ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'" ./cmd/go-read-burn
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection and coverage (CI standard)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run a single test by name
go test -v -run TestFunctionName ./path/to/package

# Run tests in a specific package
go test -v ./internal/crypto/...
```

### Linting and Static Analysis

```bash
# Run go vet
go vet ./...

# Run golangci-lint (used in CI)
golangci-lint run

# Run with specific output format
golangci-lint run --out-format checkstyle:golint-report.out
```

### Docker

```bash
# Build and run with Docker Compose
make up

# Or directly with docker compose
docker compose --project-directory deploy up --build --remove-orphans
```

## Configuration

The application reads configuration from environment variables with the `GRB_` prefix:

| Variable | Default | Description |
|----------|---------|-------------|
| `GRB_DB_PATH` | `db/secrets.db` | Path to BoltDB database file |
| `GRB_LISTEN_PORT` | `80` | HTTP server port |
| `GRB_LISTEN_HOST` | `0.0.0.0` | HTTP server bind address |

Example:
```bash
export GRB_LISTEN_HOST=127.0.0.1
export GRB_LISTEN_PORT=8080
export GRB_DB_PATH=./db/secrets.db
```

## Architecture and Code Structure

### Project Layout

```
cmd/go-read-burn/          # Main application entry point
  ├── main.go              # HTTP server setup, routing, handlers
  ├── main_test.go         # Tests for main package
  ├── views/               # HTML templates (embedded via go:embed)
  └── static/              # CSS/JS assets (embedded via go:embed)

internal/                  # Internal packages (not importable externally)
  ├── crypto/              # AES-256-GCM encryption/decryption
  │   ├── crypto.go        # Core crypto functions
  │   └── crypto_test.go   # Comprehensive crypto tests
  └── version/             # Version metadata
      └── version.go       # Version string

deploy/                    # Docker Compose configuration
db/                        # Default database directory
```

### Key Architecture Concepts

#### 1. Zero-Knowledge Encryption Model

The security model is based on **zero-knowledge encryption**:

- **ID Structure**: Each secret gets a 72-character base62 ID composed of:
  - 8 chars: Database lookup key (public)
  - 32 chars: Password for key derivation (secret)
  - 16 chars: Nonce for AES-GCM (secret)
  - 16 chars: Salt for scrypt (secret)

- **Server Ignorance**: The server stores only encrypted ciphertext using the 8-char key. The remaining 64 characters (password, nonce, salt) are given to the user and **never stored on the server**.

- **Key Derivation**: Uses scrypt with OWASP-recommended parameters (N=131072, r=8, p=1) to derive a 32-byte AES key from the password and salt.

- **Authenticated Encryption**: AES-256-GCM provides both confidentiality and integrity protection.

See `internal/crypto/crypto.go` for the complete implementation.

#### 2. Embedded Assets

The application uses Go 1.16+ `embed` directives to bundle templates and static assets directly into the binary:

```go
//go:embed all:views/*
var views embed.FS

//go:embed static/*
var static embed.FS
```

This means:
- No external file dependencies at runtime
- Single binary distribution
- Templates are parsed from the embedded filesystem

#### 3. HTTP Handlers and Routing

The application uses `gorilla/mux` for routing with three main endpoints:

- `GET /` - Index page (create secret form)
- `POST /create` - Create and store encrypted secret
- `GET /get/{key}` - Retrieve and burn secret (one-time access)
- `/static/*` - Static assets (CSS, JS)

Handler pattern in `cmd/go-read-burn/main.go`:
```go
r.HandleFunc("/", IndexHandler)
r.HandleFunc("/create", CreateHandler).Methods("POST")
r.HandleFunc("/get/{key}", SecretHandler)
```

#### 4. BoltDB Storage

BoltDB is used as an embedded key-value database:
- **Single file**: All data stored in one file (default: `db/secrets.db`)
- **No external server**: Database embedded in the application
- **ACID transactions**: Built-in transaction support
- **Read-after-write**: Secrets are deleted immediately after first read

#### 5. Graceful Shutdown

The server implements graceful shutdown:
- Listens for OS interrupt signals (Ctrl+C)
- Waits up to 30 seconds for in-flight requests to complete
- Cleanly closes database connections
- Prevents data corruption on shutdown

See `shutdownServer()` in `cmd/go-read-burn/main.go`.

## Development Guidelines

### Code Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Imports grouped: stdlib, then third-party (blank line between)
- Exported functions: PascalCase
- Unexported functions: camelCase
- Package names: lowercase, single word

### Testing Patterns

1. **Table-Driven Tests**: Preferred for testing multiple scenarios
   ```go
   tests := []struct {
       name    string
       input   string
       want    string
       wantErr bool
   }{
       // test cases...
   }
   ```

2. **Test Isolation**: Use temporary directories and files
   - Create temp DB files for database tests
   - Clean up resources in defer statements

3. **Coverage**: CI requires race detection and coverage reporting
   ```bash
   go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
   ```

### Adding New Features

When implementing new features:

1. **Security First**: This is a secret-sharing application
   - Never log secrets or encryption parameters
   - Ensure proper encryption parameter handling
   - Verify timing attack resistance
   - Follow zero-knowledge principles

2. **Test Coverage**: Write tests before or alongside implementation
   - Unit tests for internal packages
   - Integration tests for HTTP handlers using `httptest`

3. **Documentation**: Update relevant docs
   - Code comments for exported functions
   - Update AGENTS.md if architectural changes
   - Update README.md if user-facing changes

### Commit Message Format

Use **Conventional Commits** format (affects versioning):
```
<type>: <description>

Types: feat, fix, docs, test, refactor, chore, ci, build
```

Examples:
- `feat: add expiry support for secrets`
- `fix: handle missing key on /get/{key}`
- `docs: improve contributor setup instructions`
- `test: add coverage for crypto.Encrypt`

## Security Considerations

This project handles sensitive data. When making changes:

1. **Never log sensitive data**: No logging of passwords, nonces, salts, or plaintext secrets
2. **Constant-time operations**: Be aware of timing attacks in crypto operations
3. **Secure randomness**: Use `crypto/rand` for all random generation
4. **Input validation**: Validate all user inputs (ID format, length limits)
5. **OWASP guidelines**: Follow OWASP recommendations for crypto parameters

Security issues should be reported per [SECURITY.md](../../SECURITY.md), not via public issues.

## Common Tasks

### Adding a New HTTP Endpoint

1. Add route in `setupRoutes()` in `main.go`
2. Implement handler function following existing patterns
3. Add tests in `main_test.go` using `httptest`
4. Update any relevant templates in `views/`

### Modifying Encryption Logic

1. Make changes in `internal/crypto/crypto.go`
2. Update tests in `internal/crypto/crypto_test.go`
3. Run full test suite with race detection
4. Consider backward compatibility with existing stored secrets

### Adding Configuration Options

1. Add field to `Config` struct in `main.go`
2. Use envconfig struct tags for environment variable mapping
3. Document in README.md and this file
4. Provide sensible defaults

## CI/CD Pipeline

GitHub Actions runs on all PRs and pushes:
- **Build**: Compiles all packages
- **Test**: Runs tests with race detection and coverage
- **Lint**: Runs golangci-lint
- **SonarCloud**: Code quality and security analysis
- **Version**: Uses GitVersion for semantic versioning

All jobs must pass before merge.

## Useful References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [OWASP Cryptographic Storage](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Contributing Guide](../../CONTRIBUTING.md)
