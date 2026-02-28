# Development Guide

This guide covers how to set up a local development environment, run the
application, and work effectively with the codebase.

---

## Prerequisites

| Tool               | Minimum version | Install                               |
|--------------------|-----------------|---------------------------------------|
| Go                 | 1.25            | <https://go.dev/dl/>                  |
| golangci-lint      | 2.x             | <https://golangci-lint.run/usage/install/> |
| Git                | 2.x             | system package manager                |
| `gh` CLI (optional)| any             | <https://cli.github.com/>             |

The project uses `modernc.org/sqlite` — a pure-Go SQLite driver — so **no C
toolchain (CGo) is required**.

---

## Clone and Run

```bash
git clone git@github.com:gjcourt/drift.git
cd drift
make build       # compile ./cmd/drift → ./drift
make test        # run all tests with -race
./drift          # start the server on :8080
```

Open <http://localhost:8080> in your browser.

---

## Environment Variables

All configuration is supplied via environment variables. Defaults are suitable
for local development when running from the repository root.

| Variable          | Default (local dev)                                            | Description                                              |
|-------------------|----------------------------------------------------------------|----------------------------------------------------------|
| `DRIFT_ADDR`      | `:8080`                                                        | TCP address the HTTP server listens on                   |
| `DRIFT_DB`        | `drift.db`                                                     | Path to the SQLite database file. Created on first run.  |
| `DRIFT_TMPL_DIR`  | `<repo>/internal/adapters/http/templates` (resolved at build) | Directory containing Go `html/template` files            |
| `DRIFT_STATIC_DIR`| `<repo>/web/static` (resolved at build)                       | Directory served at `/static/`                           |

`DRIFT_TMPL_DIR` and `DRIFT_STATIC_DIR` are resolved relative to the source
file location at build time (via `runtime.Caller`). Override them in
containerised or non-standard deployments.

**Example:**

```bash
DRIFT_ADDR=:9090 DRIFT_DB=/data/drift.db ./drift
```

---

## Makefile Targets

Run `make <target>` from the repository root.

| Target   | Command(s)                                           | Description                                   |
|----------|------------------------------------------------------|-----------------------------------------------|
| `build`  | `go build -o drift ./cmd/drift`                      | Compile the binary                            |
| `run`    | `go run ./cmd/drift`                                 | Run without compiling to disk                 |
| `dev`    | `go run ./cmd/drift` with auto-restart (if watcher)  | Development shortcut                          |
| `test`   | `go test -race -count=1 ./...`                       | Run all tests with the race detector          |
| `fmt`    | `go fmt ./...`                                       | Format all Go source files                    |
| `vet`    | `go vet ./...`                                       | Run the Go static analyser                    |
| `lint`   | `golangci-lint run ./...`                            | Run all configured linters                    |
| `check`  | `fmt` + `vet` + `lint` + `test`                      | Full quality gate — run before pushing        |
| `tidy`   | `go mod tidy && go mod verify`                       | Tidy and verify the module graph              |
| `clean`  | `rm -f drift`                                        | Remove the compiled binary                    |

Run `make check` before opening a pull request.

---

## Testing

### Running Tests

```bash
make test                            # all packages, race detector
go test -race ./internal/services/... # single package
go test -run TestUploadCSV ./...      # single test
```

### Coverage

```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Conventions

- **Table-driven**: use `[]struct{ name, input, want }` subtests with `t.Run`.
- **SQLite in-memory**: open stores with `sqlite.New(":memory:")` — no files,
  no cleanup needed.
- **Race detector**: always pass `-race`; the CI pipeline enforces this.
- **`errcheck` suppressed in tests**: `//nolint:errcheck` is allowed on
  deferred cleanup calls inside `_test.go` files (see `.golangci.yml`).

### Test Files

| File                                                  | Package under test           |
|-------------------------------------------------------|------------------------------|
| `internal/domain/domain_test.go`                      | domain types & validation    |
| `internal/adapters/ingestion/csv_test.go`             | CSV parser                   |
| `internal/adapters/ingestion/json_test.go`            | JSON config parser           |
| `internal/adapters/storage/sqlite/db_test.go`         | SQLite store                 |
| `internal/services/ingestion_test.go`                 | ingestion service            |
| `internal/services/simulation_test.go`                | simulation service           |
| `internal/services/results_test.go`                   | results service              |

---

## Linting

Drift uses [golangci-lint](https://golangci-lint.run/) v2. The configuration is
in [`.golangci.yml`](../.golangci.yml).

Enabled linters (beyond the `standard` preset):

| Linter      | Purpose                                       |
|-------------|-----------------------------------------------|
| `revive`    | Opinionated Go style (exported docs, naming)  |
| `misspell`  | Common English spelling errors in comments    |
| `gocritic`  | Diagnostic and style checks                   |
| `gofmt`     | Canonical formatting (formatter)              |
| `goimports` | Import ordering (formatter)                   |

Run the linter:

```bash
make lint          # golangci-lint run ./...
golangci-lint run --fix ./...   # apply auto-fixes where possible
```

---

## Branch Workflow

See [CONTRIBUTING.md](../CONTRIBUTING.md) for the full contribution guide,
including branch naming conventions, commit format, and PR requirements.

Quick summary:

1. Branch from `main`: `git checkout -b feat/my-feature`
2. Make changes, run `make check`
3. Push and open a PR against `main`
4. All six CI jobs must pass; one approval required
5. Squash-merge only

---

## Project Layout

```
cmd/drift/          # main package — wires everything together
internal/
  domain/           # core types (no dependencies)
  ports/
    inbound/        # service interfaces called by HTTP handlers
    outbound/       # repository interfaces called by services
  services/         # business logic (depends on ports only)
  adapters/
    http/           # chi router, handlers, templates
    ingestion/      # CSV and JSON parsers
    storage/sqlite/ # SQLite repository implementation
web/static/         # JS, CSS, and vendor assets served at /static/
docs/               # project documentation (you are here)
```

For a deeper architectural overview see [architecture.md](architecture.md).
