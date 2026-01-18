# go-read-burn

[![Open in Visual Studio Code](https://img.shields.io/static/v1?logo=visualstudiocode&label=&message=Open%20in%20Visual%20Studio%20Code&labelColor=2c2c32&color=007acc&logoColor=007acc)](https://open.vscode.dev/danstis/go-read-burn)
[![Go Report Card](https://goreportcard.com/badge/github.com/danstis/go-read-burn?style=flat-square)](https://goreportcard.com/report/github.com/danstis/go-read-burn)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/danstis/go-read-burn)](https://pkg.go.dev/github.com/danstis/go-read-burn)
[![Release](https://img.shields.io/github/release/danstis/go-read-burn.svg?style=flat-square)](https://github.com/danstis/go-read-burn/releases/latest)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danstis_go-read-burn&metric=alert_status)](https://sonarcloud.io/dashboard?id=danstis_go-read-burn)

go-read-burn is a shameless re-creation of the fantastic Node.JS app [Read2Burn](https://www.read2burn.com/) by Wemove, written in Go.

## Contributing

Contributions are welcome — bug reports, feature requests, documentation improvements, and code changes.

### Quick links
- Issues: https://github.com/danstis/go-read-burn/issues
- Pull Requests: https://github.com/danstis/go-read-burn/pulls

### Fork & pull request workflow (required)
This project expects contributions to come via the standard GitHub fork + PR flow.

1. Fork the repo.
2. Clone your fork.
3. Create a branch (`feature/...` or `fix/...`).
4. Make changes and add/update tests where it makes sense.
5. Push to your fork and open a PR back to this repo.

### Development setup (local)

**Requirements**
- Go **1.20+** (see `go.mod`)
- (Optional) Docker + Docker Compose
- (Optional) `gitversion` if you want `make run` to include version metadata

**Run locally**
```bash
go run ./cmd/go-read-burn
```

**Run with version info (uses gitversion)**
```bash
make run
```

**Run tests**
```bash
go test ./...
```

**Run tests like CI (race + coverage)**
```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

**Static analysis**
```bash
go vet ./...
```

**Lint**
```bash
golangci-lint run
```

### Conventions
- Project layout follows a standard Go layout (see `cmd/` and `internal/`).
- Commit messages follow **Conventional Commits** (this impacts versioning/changelog):
  https://www.conventionalcommits.org/

### Community & security
- Code of Conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Security reporting: [SECURITY.md](SECURITY.md) (tracking issue: #126)

For more details, see [CONTRIBUTING.md](CONTRIBUTING.md).

