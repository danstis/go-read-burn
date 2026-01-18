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

## Test-Driven Development (TDD) Methodology

### TDD Philosophy

This project follows **Test-Driven Development (TDD)** as the **highly recommended default** approach for all new features and bug fixes. TDD helps ensure code correctness, maintainability, and provides living documentation of expected behavior.

**When to use TDD**:

- ✅ **Default approach**: All new features, handlers, and business logic
- ✅ Bug fixes (write a failing test that reproduces the bug, then fix it)
- ✅ Refactoring existing code
- ✅ API endpoint implementation
- ✅ Database operations and data validation

**When TDD may be skipped** (with clear justification required):

- Exploratory spikes or throwaway prototypes
- Emergency production hotfixes (add tests immediately after)
- Simple refactoring with existing test coverage
- Pure visual/styling changes without logic

> **Important**: If not using TDD, document the justification in commit messages, PR descriptions, or code comments.

### The Red-Green-Refactor Cycle

TDD follows a three-phase cycle:

```
RED → GREEN → REFACTOR → RED → GREEN → REFACTOR → ...
```

#### 🔴 RED Phase: Write a Failing Test

1. Write a test for the desired behavior
2. The test should fail for the **right reason** (not compilation errors)
3. Run the test to confirm it fails: `go test -v ./...`

**Why this matters**: Confirms your test actually validates the behavior and can detect failures.

#### 🟢 GREEN Phase: Make the Test Pass

1. Write the **minimum code** necessary to make the test pass
2. Don't worry about perfection or optimization yet
3. Run tests to confirm they pass: `go test -v ./...`

**Why this matters**: Focuses on solving the immediate problem without over-engineering.

#### 🔵 REFACTOR Phase: Improve the Code

1. Improve code quality, readability, and structure
2. Apply Go best practices and project conventions
3. Keep tests passing throughout refactoring
4. Run tests frequently: `go test -v ./...`

**Why this matters**: Ensures high-quality code while tests provide a safety net.

### TDD Process: Step-by-Step

1. **Understand the requirement**: Know what you're building
2. **Write a descriptive test name**: `TestCreateHandler_ValidSecret_ReturnsCreated`
3. **Write the test implementation**: Test should fail initially
4. **Run tests and watch it fail**: `go test -v ./...` (RED)
5. **Write minimal production code**: Just enough to pass
6. **Run tests and watch them pass**: `go test -v ./...` (GREEN)
7. **Refactor if needed**: Improve while keeping tests green (REFACTOR)
8. **Commit with tests**: Commit implementation and tests together

### Practical TDD Examples

#### Example 1: HTTP Handler with Table-Driven Tests

Table-driven tests are ideal for testing multiple scenarios with different inputs.

```go
// Step 1: Write the test FIRST (in main_test.go)
func TestCreateHandler(t *testing.T) {
    // Initialize test database
    tempDir := t.TempDir()
    testDB, err := bolt.Open(filepath.Join(tempDir, "test.db"), 0644, nil)
    if err != nil {
        t.Fatalf("Failed to open test DB: %v", err)
    }
    defer testDB.Close()
    db = testDB // Set global db for handler

    tests := []struct {
        name           string
        requestBody    string
        expectedStatus int
        checkBody      func(t *testing.T, body string)
    }{
        {
            name:           "valid secret creation",
            requestBody:    `{"secret":"test secret","expiry":3600}`,
            expectedStatus: http.StatusCreated,
            checkBody: func(t *testing.T, body string) {
                if !strings.Contains(body, `"key":`) {
                    t.Error("Response should contain a key")
                }
            },
        },
        {
            name:           "empty secret",
            requestBody:    `{"secret":"","expiry":3600}`,
            expectedStatus: http.StatusBadRequest,
            checkBody: func(t *testing.T, body string) {
                if !strings.Contains(body, "error") {
                    t.Error("Response should contain error message")
                }
            },
        },
        {
            name:           "invalid json",
            requestBody:    `{invalid}`,
            expectedStatus: http.StatusBadRequest,
            checkBody:      nil, // Optional body check
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/create",
                bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            rr := httptest.NewRecorder()

            CreateHandler(rr, req)

            if rr.Code != tt.expectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v",
                    rr.Code, tt.expectedStatus)
            }

            if tt.checkBody != nil {
                tt.checkBody(t, rr.Body.String())
            }
        })
    }
}

// Step 2: Run test - it WILL FAIL (RED)
// $ go test -v -run TestCreateHandler ./cmd/go-read-burn
// --- FAIL: TestCreateHandler (0.00s)
//     main_test.go:XX: handler returned wrong status code: got 200 want 201

// Step 3: Implement CreateHandler to make it pass (GREEN)
func CreateHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Secret string `json:"secret"`
        Expiry int    `json:"expiry"`
    }

    // Decode request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate
    if req.Secret == "" {
        http.Error(w, "Secret cannot be empty", http.StatusBadRequest)
        return
    }

    // Save to database (implement SaveSecret function)
    key, err := SaveSecret(db, req.Secret, req.Expiry)
    if err != nil {
        http.Error(w, "Failed to save secret", http.StatusInternalServerError)
        return
    }

    // Return success
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"key": key})
}

// Step 4: Run tests again - they should PASS (GREEN)
// $ go test -v -run TestCreateHandler ./cmd/go-read-burn
// --- PASS: TestCreateHandler (0.01s)

// Step 5: Refactor if needed (REFACTOR)
// - Extract validation logic
// - Improve error messages
// - Add more edge case tests
```

#### Example 2: Database Operations with Temporary DB

Always use temporary databases for testing to ensure isolation.

```go
func TestSaveSecret(t *testing.T) {
    // Arrange: Create temporary database
    tempDir := t.TempDir() // Automatically cleaned up
    db, err := bolt.Open(filepath.Join(tempDir, "test.db"), 0644, nil)
    if err != nil {
        t.Fatalf("Failed to open test DB: %v", err)
    }
    defer db.Close()

    // Create bucket for testing
    err = db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte("secrets"))
        return err
    })
    if err != nil {
        t.Fatalf("Failed to create bucket: %v", err)
    }

    tests := []struct {
        name    string
        secret  string
        expiry  int
        wantErr bool
    }{
        {
            name:    "valid secret",
            secret:  "my secret data",
            expiry:  3600,
            wantErr: false,
        },
        {
            name:    "empty secret",
            secret:  "",
            expiry:  3600,
            wantErr: true,
        },
        {
            name:    "zero expiry",
            secret:  "test",
            expiry:  0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Act: Save the secret
            key, err := SaveSecret(db, tt.secret, tt.expiry)

            // Assert: Check error condition
            if (err != nil) != tt.wantErr {
                t.Errorf("SaveSecret() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr {
                return // Don't check further if we expected an error
            }

            // Assert: Verify key was generated
            if key == "" {
                t.Error("Expected non-empty key")
            }

            // Assert: Verify it was actually saved
            retrieved, err := GetSecret(db, key)
            if err != nil {
                t.Errorf("Failed to retrieve secret: %v", err)
            }
            if retrieved != tt.secret {
                t.Errorf("Retrieved secret = %v, want %v", retrieved, tt.secret)
            }
        })
    }
}
```

#### Example 3: Configuration Validation

Test environment variable loading and defaults.

```go
func TestLoadConfig(t *testing.T) {
    tests := []struct {
        name       string
        envVars    map[string]string
        wantErr    bool
        wantConfig Config
    }{
        {
            name:    "default values",
            envVars: map[string]string{},
            wantErr: false,
            wantConfig: Config{
                DBPath:     "db/secrets.db",
                ListenPort: "80",
                ListenHost: "0.0.0.0",
            },
        },
        {
            name: "custom values",
            envVars: map[string]string{
                "GRB_DB_PATH":     "/custom/path.db",
                "GRB_LISTEN_PORT": "8080",
                "GRB_LISTEN_HOST": "127.0.0.1",
            },
            wantErr: false,
            wantConfig: Config{
                DBPath:     "/custom/path.db",
                ListenPort: "8080",
                ListenHost: "127.0.0.1",
            },
        },
        {
            name: "partial custom values",
            envVars: map[string]string{
                "GRB_LISTEN_PORT": "3000",
            },
            wantErr: false,
            wantConfig: Config{
                DBPath:     "db/secrets.db",
                ListenPort: "3000",
                ListenHost: "0.0.0.0",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange: Set environment variables
            for k, v := range tt.envVars {
                t.Setenv(k, v) // Automatically restored after test
            }

            // Act: Load configuration
            config, err := loadConfig()

            // Assert: Check error
            if (err != nil) != tt.wantErr {
                t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // Assert: Check configuration values
            if config.DBPath != tt.wantConfig.DBPath {
                t.Errorf("DBPath = %v, want %v", config.DBPath, tt.wantConfig.DBPath)
            }
            if config.ListenPort != tt.wantConfig.ListenPort {
                t.Errorf("ListenPort = %v, want %v", config.ListenPort, tt.wantConfig.ListenPort)
            }
            if config.ListenHost != tt.wantConfig.ListenHost {
                t.Errorf("ListenHost = %v, want %v", config.ListenHost, tt.wantConfig.ListenHost)
            }
        })
    }
}
```

### Testing Best Practices

#### Test Organization

- **One test file per source file**: `main.go` → `main_test.go`
- **Group related tests**: Use subtests with `t.Run()` for related scenarios
- **Table-driven tests**: Use for testing multiple inputs/outputs
- **Test both success and error paths**: Don't just test happy paths

#### Test Naming Conventions

```go
// Basic test
func TestFunctionName(t *testing.T) { }

// Specific scenario test
func TestFunctionName_SpecificScenario(t *testing.T) { }

// Examples
func TestCreateHandler(t *testing.T) { }
func TestCreateHandler_EmptySecret_ReturnsBadRequest(t *testing.T) { }
func TestSaveSecret_DatabaseError_ReturnsError(t *testing.T) { }

// Descriptive subtest names
t.Run("empty secret returns error", func(t *testing.T) { })
t.Run("valid secret creates key", func(t *testing.T) { })
```

#### Test Structure: Arrange-Act-Assert (AAA)

```go
func TestExample(t *testing.T) {
    // Arrange: Setup test data and dependencies
    input := "test"
    expected := "expected"
    mock := setupMockDB(t)

    // Act: Execute the code under test
    result, err := DoSomething(mock, input)

    // Assert: Verify the results
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

#### Test Isolation

**Use Go testing helpers for automatic cleanup**:

```go
// Temporary directories (auto-cleanup)
tempDir := t.TempDir()
dbPath := filepath.Join(tempDir, "test.db")

// Environment variables (auto-restore)
t.Setenv("GRB_DB_PATH", "/custom/path")

// Manual cleanup when needed
cleanup := setupTestData()
defer cleanup()
```

**Create fresh test databases**:

```go
func setupTestDB(t *testing.T) *bolt.DB {
    t.Helper() // Marks this as a helper function

    tempDir := t.TempDir()
    db, err := bolt.Open(filepath.Join(tempDir, "test.db"), 0644, nil)
    if err != nil {
        t.Fatalf("Failed to open test DB: %v", err)
    }

    // Initialize required buckets
    err = db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte("secrets"))
        return err
    })
    if err != nil {
        t.Fatalf("Failed to initialize DB: %v", err)
    }

    t.Cleanup(func() { db.Close() })
    return db
}
```

#### Error Path Testing

Always test error conditions explicitly:

```go
func TestOpenDB_InvalidPath_ReturnsError(t *testing.T) {
    // Test with path that will fail
    invalidPath := "/root/cannot-write-here/db.db"

    _, err := openDB(invalidPath)

    if err == nil {
        t.Error("Expected error for invalid path, got nil")
    }

    // Verify error message is helpful
    if !strings.Contains(err.Error(), "failed to create database directory") {
        t.Errorf("Error message not descriptive: %v", err)
    }
}
```

#### Using Helper Functions

Mark helper functions with `t.Helper()`:

```go
func assertStatusCode(t *testing.T, got, want int) {
    t.Helper() // Test failures report caller's line, not this line

    if got != want {
        t.Errorf("status code = %v, want %v", got, want)
    }
}

func TestSomething(t *testing.T) {
    // ...
    assertStatusCode(t, rr.Code, http.StatusOK) // Failure reports THIS line
}
```

### Test Coverage Requirements

#### Coverage Targets

| Area | Target | Rationale |
|------|--------|-----------|
| **Overall Project** | 60% minimum | Baseline quality standard |
| **HTTP Handlers** | 80%+ | Critical user-facing code |
| **Database Operations** | 80%+ | Data integrity critical |
| **Utility Functions** | 60%+ | Supporting code |

#### What to Test

- ✅ **Business logic and algorithms**
- ✅ **HTTP handlers** (request validation, response formatting)
- ✅ **Database operations** (CRUD, transactions)
- ✅ **Error handling and validation**
- ✅ **Configuration loading** (env vars, defaults)
- ✅ **Edge cases and boundary conditions**

#### What NOT to Test (typically)

- ❌ **Third-party library internals** (trust they're tested)
- ❌ **Standard library functions** (e.g., `strings.Contains`)
- ❌ **Simple getters/setters** without logic
- ❌ **Generated code**

#### Measuring Coverage

```bash
# Generate coverage report
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# View coverage by function
go tool cover -func=coverage.out

# View HTML coverage report (opens in browser)
go tool cover -html=coverage.out -o coverage.html

# Check overall coverage percentage
go test -cover ./...

# Example output:
# ?       github.com/danstis/go-read-burn/internal/version   [no test files]
# ok      github.com/danstis/go-read-burn/cmd/go-read-burn   0.234s  coverage: 65.4% of statements
```

### Running Tests During Development

#### Continuous Testing Workflow

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test -v ./cmd/go-read-burn

# Run a specific test by name
go test -v -run TestCreateHandler ./cmd/go-read-burn

# Run tests matching a pattern
go test -v -run "TestCreate.*" ./...

# Run tests with race detection (recommended)
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Full test suite (as run in CI)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

#### Pre-Commit Testing Checklist

Before committing code, always run:

```bash
# 1. Run all tests
go test ./...

# 2. Check coverage meets minimum (60%)
go test -cover ./...

# 3. Run race detector
go test -race ./...

# 4. Run linter
go vet ./...

# 5. Run golangci-lint (if available)
golangci-lint run
```

### TDD with Existing Code

When working with existing code that lacks tests:

1. **Write characterization tests**: Document current behavior
2. **Then refactor safely**: Tests catch regressions
3. **Gradually increase coverage**: Don't try to test everything at once

**Example workflow**:

```go
// 1. Write test for current behavior (even if imperfect)
func TestExistingFunction_CurrentBehavior(t *testing.T) {
    result := ExistingFunction("input")

    // Document what it currently does
    if result != "current-output" {
        t.Errorf("Behavior changed: got %v", result)
    }
}

// 2. Now you can safely refactor
// 3. Add more tests for edge cases
// 4. Improve implementation with confidence
```

### Justification for Non-TDD Approaches

When **NOT** using TDD, document the reason clearly:

#### Valid Exceptions

| Exception | When to Use |
|-----------|-------------|
| **Exploratory spikes** | Throwaway code to understand a problem/API |
| **Rapid prototyping** | Proof-of-concept that will be rewritten |
| **Simple refactoring** | Moving code without changing behavior (existing tests cover it) |
| **Emergency hotfixes** | Critical production issues requiring immediate fix |
| **Visual/UI adjustments** | Pure styling changes with no logic changes |

#### Documentation Requirements

**In commit messages**:

```
fix: emergency hotfix for database connection leak

TDD not used due to production emergency requiring immediate fix.
Follow-up issue #123 created to add comprehensive tests.

Closes #122
```

**In PR descriptions**:

```markdown
## Testing Approach

This PR uses exploratory spike code to evaluate the new authentication library.
TDD was not used because this is throwaway code for evaluation only.

Once we decide on the approach, a follow-up PR will implement the feature
properly using TDD methodology.
```

### TDD Benefits for This Project

| Benefit | Description |
|---------|-------------|
| **Confidence in changes** | Refactor without fear of breaking things |
| **Better API design** | Writing tests first reveals awkward APIs early |
| **Living documentation** | Tests show how code should be used |
| **Faster debugging** | Tests pinpoint exactly what broke |
| **Easier code review** | Tests demonstrate intent clearly |
| **Regression prevention** | Bugs stay fixed |

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

### Build and Test Workflow

The CI/CD pipeline enforces code quality through multiple stages:

1. **Version Generation**: GitVersion calculates semantic version
2. **Build**: `go build -v ./...` - Must pass for workflow to succeed
3. **Test**: `go test -v -race -coverprofile=coverage.out -covermode=atomic ./...`
   - Tests continue on error (`continue-on-error: true`) to collect coverage artifacts
   - Coverage reports and test output uploaded as artifacts
   - Reports sent to SonarCloud for quality analysis
   - **Overall workflow fails if tests fail** (after artifact collection)
   - **Minimum coverage**: 60%
4. **Lint**: golangci-lint with checkstyle output
   - Continues on error to collect lint reports
   - Reports uploaded and sent to SonarCloud
5. **Quality Gate**: SonarCloud analysis
   - Evaluates code quality, coverage, and technical debt
   - Quality gate must pass for main branch merges
6. **Release**: GoReleaser (triggered automatically on main branch)

### Test Failure Policy

**Why `continue-on-error: true`?**

Tests and linting use `continue-on-error: true` to ensure diagnostic artifacts are generated even when failures occur. This allows:

- Coverage reports to be uploaded to SonarCloud
- Test output to be available for debugging
- Lint reports to be analyzed for quality metrics

**However**: The overall workflow **will fail** if tests fail, maintaining our quality gates while ensuring we have complete diagnostic information.

### Local Testing Before Push

Always run the full test suite locally before pushing:

```bash
# Run all checks (as executed in CI)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Check coverage meets minimum (60%)
go tool cover -func=coverage.out

# Run linters
go vet ./...
golangci-lint run

# Verify build succeeds
go build -v ./...
```

### Coverage Monitoring

- **SonarCloud** tracks coverage trends over time
- **Minimum**: 60% overall coverage required
- **New code**: Should maintain or improve coverage percentage
- Review coverage reports in PR checks before merging

### Quality Gates

Before merging to `main`:

- ✅ All tests must pass
- ✅ Build must succeed
- ✅ Coverage must be ≥60%
- ✅ SonarCloud quality gate must pass
- ✅ Code review approval required

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
