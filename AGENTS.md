# AGENTS.md

> Drift is a Monte Carlo stock portfolio simulation engine in Go — runs GBM and bootstrapped-returns simulations over multi-asset portfolios with a self-hosted web UI. — https://github.com/gjcourt/drift

## Commands

| Command | Use |
|---------|-----|
| `make dev` | Run with hot-reload-friendly flags (`DRIFT_ADDR=:8080`) |
| `make build` | Compile binary to `./drift` |
| `make run` | Build + run |
| `make test` | Run unit tests with race detector |
| `make fmt` | gofmt |
| `make vet` | go vet |
| `make lint` | golangci-lint |
| `make check` | fmt + vet + lint + test |
| `make tidy` | `go mod tidy` + verify |

Single test: `go test ./internal/app -run TestSimulation -v`
Pre-push: `make check`

## Architecture

Hexagonal architecture (ports & adapters). Entry point: `cmd/drift/main.go`.

- `internal/domain/` — entity types: assets, paths, portfolios, experiments, results. No external deps.
- `internal/app/` — use-case orchestration (simulation, ingestion, results); depends on `domain/` and `ports/...`.
- `internal/ports/inbound/` — interfaces the app exposes (results, ingestion, simulation).
- `internal/ports/outbound/` — interfaces the app requires (storage, CSV parsing).
- `internal/adapters/http/` — driving adapter (HTTP server + handlers + templates).
- `internal/adapters/ingestion/` — CSV / JSON ingestion; implements `outbound.CSVParser`.
- `internal/adapters/storage/sqlite/` — SQLite persistence; implements all repository ports.
- `web/` — HTML templates and frontend assets.

See `docs/architecture/` for the full guide.

## Conventions

- **Simulation logic lives in `app/` and `domain/`** — never in HTTP handlers.
- **`internal/app/` depends only on `domain/` and `ports/...`** — never on adapters.
- **Adapters implement ports** — adapters translate between external formats and domain types.
- **No ORM** — SQLite is accessed via stdlib `database/sql` only.
- **Templates** are resolved from source tree in dev (`DRIFT_TMPL_DIR` set) and embedded in the binary for production.
- **Conventional Commits** for every commit (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, `test:`, `ci:`).
- **Branch names** follow `<type>/<description>`.
- **Every exported symbol** has a Go doc comment. No bare `//nolint` — every suppression names the linter and explains why.

## Invariants

- `internal/domain/` must not import any third-party packages outside stdlib.
- `internal/ports/` must not import `internal/adapters/` or `internal/app/`.
- `internal/app/` must not import `internal/adapters/`.
- `internal/adapters/http/handlers/` translates request → port → response only — no simulation logic.
- The local SQLite file `drift.db` is gitignored and never committed.

## What NOT to Do

- Do not put SQL or HTTP types in `internal/app/` or `internal/domain/` — adapters translate, core stays pure.
- Do not import `internal/adapters/` from `internal/ports/` or `internal/app/`.
- Do not skip `make check` before committing — formatting / vet / lint / test must all be green.
- Do not commit `drift.db` or any uploaded CSV under `data/`.

## Domain

Drift simulates the future value of a stock portfolio by Monte Carlo: from historical price data the user uploads, it estimates per-asset return distributions, then runs many forward paths (GBM or bootstrap) over a user-specified horizon. The web UI surfaces fan charts, percentile bands (p5/p25/p50/p75/p95), drawdown distributions, probability of loss, and CAGR statistics.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `DRIFT_ADDR` | `:8080` | HTTP listen address |
| `DRIFT_DB` | `drift.db` | SQLite database file path |
| `DRIFT_TMPL_DIR` | (auto) | Path to HTML template directory; auto-detected in dev, embedded in prod |

## Cross-service dependencies

_n/a — Drift is self-contained: a single Go binary plus an embedded SQLite file. No external services._

## Quality gate before push

1. `make fmt`
2. `make vet`
3. `make lint`
4. `make test`

Or `make check`, which runs them in order.

## Documentation

`docs/` taxonomy: `architecture/` · `design/` · `operations/` · `plans/` · `reference/` · `research/`. See each folder's `README.md` for scope. Index: `docs/README.md`.

## Observability

Logs to stderr in slog text format. No metrics endpoint today — the web UI's experiment list is the canonical health surface for whether simulations are completing.

When you learn a new convention or invariant in this repo, update this file.
