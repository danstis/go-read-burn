# Contributing to go-read-burn

Thanks for your interest in contributing! This document describes how to set up a dev environment, propose changes, and get PRs merged smoothly.

## Table of contents
- [Ways to contribute](#ways-to-contribute)
- [Fork & PR workflow (required)](#fork--pr-workflow-required)
- [Development setup](#development-setup)
- [Running locally](#running-locally)
- [Testing](#testing)
- [Linting & formatting](#linting--formatting)
- [Project structure](#project-structure)
- [Commit messages (Conventional Commits)](#commit-messages-conventional-commits)
- [Pull request expectations](#pull-request-expectations)
- [Security](#security)

## Ways to contribute

You can contribute by:
- Reporting bugs (include repro steps and logs)
- Suggesting features / UX improvements
- Improving documentation (README, comments)
- Adding tests or improving CI/dev experience
- Implementing code changes

If you’re unsure where to start, open an issue first describing what you want to do.

## Fork & PR workflow (required)

This project expects contributions to come via the standard GitHub fork + pull request flow.

1. **Fork** the repository on GitHub.
2. **Clone** your fork locally.
3. Create a **branch** for your change:
   - `feature/<short-description>` for new features
   - `fix/<short-description>` for bug fixes
4. Make your change locally.
5. Push the branch to your fork.
6. Open a **Pull Request** (PR) back to this repository.

## Development setup

### Requirements
- Go **1.20+** (see `go.mod`)
- Git
- Optional:
  - Docker + Docker Compose
  - `gitversion` (used by `make run` to inject version/build metadata)

### Configuration / environment variables
The app reads configuration from environment variables using the `GRB_` prefix.

| Variable | Default | Description |
|----------|---------|-------------|
| `GRB_DB_PATH` | `db/secrets.db` | Path to BoltDB database file |
| `GRB_LISTEN_PORT` | `80` | HTTP server port |
| `GRB_LISTEN_HOST` | `0.0.0.0` | HTTP server host |

Example:

```bash
export GRB_LISTEN_HOST=127.0.0.1
export GRB_LISTEN_PORT=8080
export GRB_DB_PATH=./db/secrets.db
```

## Running locally

### Run directly with Go

```bash
go run ./cmd/go-read-burn
```

### Run via Make (includes version metadata)

The repo provides a `makefile` target:

```bash
make run
```

Note: `make run` calls `gitversion` to compute `VERSION`. If you don’t have `gitversion` installed, run via `go run` as shown above.

### Run with Docker Compose

```bash
make up
# or:
docker compose --project-directory deploy up --build --remove-orphans
```

## Testing

### Run all tests

```bash
go test ./...
```

### Run tests like CI (race + coverage)

CI runs race detection and produces coverage output:

```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

### Writing tests

Preferred patterns in this repo:
- Use Go’s standard `testing` package.
- Use `httptest` for HTTP handler tests.
- Prefer table-driven tests where it improves clarity.
- Keep tests isolated (temp dirs, temporary DB files, etc.).

If you’re fixing a bug, a good workflow is:
1. Add a failing test that reproduces the bug
2. Fix the bug
3. Keep the change minimal and focused

## Linting & formatting

### Format

```bash
go fmt ./...
```

### Vet

```bash
go vet ./...
```

### golangci-lint

CI uses golangci-lint. Run locally:

```bash
golangci-lint run
```

Editor/formatting conventions are also captured in `.editorconfig`.

## Project structure

This repo follows a standard Go project layout:
- `cmd/go-read-burn/` — main application entrypoint (HTTP server, handlers, embedded templates/static assets)
- `internal/` — internal packages (not meant for external import)
- `deploy/` — deployment configs (compose, env, etc.)
- `db/` — local db directory (BoltDB file path defaults under here)

## Commit messages (Conventional Commits)

This repository uses **Conventional Commits** so versioning/changelog automation works.

Examples:
- `feat: add expiry support for secrets`
- `fix: handle missing key on /get/{key}`
- `docs: improve contributor setup instructions`
- `test: add coverage for IndexHandler`
- `ci: update Go version in workflows`

Keep commits focused; avoid mixing refactors with bug fixes unless necessary.

## Pull request expectations

Before opening a PR, please:
- Ensure tests pass: `go test ./...`
- Ideally also run: `go vet ./...` and `golangci-lint run`

In your PR description, include:
- What you changed and why
- How to test it
- Any tradeoffs or follow-ups

## Security

If you discover a security issue (anything that exposes secrets, makes them retrievable after “burn”, etc.), please follow the instructions in [SECURITY.md](SECURITY.md).
