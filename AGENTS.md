# Drift Agent Guidelines

## Repository Overview

Drift is a Monte Carlo stock portfolio simulation engine written in Go. It runs Geometric Brownian Motion (GBM) and bootstrapped-returns simulations over multi-asset portfolios, and visualises fan charts, percentile bands, drawdown distributions, and summary statistics through a self-hosted web UI.

## Project Structure

```
cmd/drift/         ← entry point
internal/          ← simulation engine, domain logic, HTTP handlers
web/               ← HTML templates and frontend assets
docs/              ← architecture and usage documentation
drift.db           ← local SQLite database (dev)
```

## Common Commands

```bash
make dev           # run with hot-reload-friendly flags (DRIFT_ADDR=:8080)
make build         # compile binary
make run           # build and run
make test          # run unit tests with race detector
make lint          # run golangci-lint
make fmt           # run gofmt
make check         # fmt + vet + lint + test
make tidy          # go mod tidy
```

## Environment Variables

| Variable         | Default      | Description                       |
|------------------|--------------|-----------------------------------|
| `DRIFT_ADDR`     | `:8080`      | HTTP listen address               |
| `DRIFT_DB`       | `drift.db`   | SQLite database file path         |
| `DRIFT_TMPL_DIR` | (auto)       | Path to HTML template directory   |

## Architecture Guidelines

- Simulation logic lives in `internal/` — keep it decoupled from HTTP handlers.
- Database is SQLite via the standard `database/sql` interface — no ORM.
- Templates are resolved from source tree in dev mode; embedded in binary for production builds.
- Run `make check` before committing.

## Usage Notes

1. Upload price data at `/data` — single-symbol (`AAPL.csv`) or multi-symbol CSV with a `symbol` column.
2. Create experiments at `/experiments/new` — select assets, weights, and simulation parameters.
3. View results: percentile distribution charts, p5/p25/p50/p75/p95, probability of loss, CAGR, max drawdown.
