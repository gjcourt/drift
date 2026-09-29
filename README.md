# Drift

> Monte Carlo stock portfolio simulation engine.

Drift runs Geometric Brownian Motion (GBM) and bootstrapped-returns simulations over
multi-asset portfolios and visualises fan charts, percentile bands, drawdown distributions,
and summary statistics through a self-hosted web UI.

## Quick Start

```bash
# 1. Install dependencies
go mod tidy

# 2. Run (templates resolved from source tree)
make dev
# Open http://localhost:8080

# Or build a binary
make build
./drift
```

## Environment Variables

| Variable          | Default      | Description                       |
|-------------------|--------------|-----------------------------------|
| `DRIFT_ADDR`      | `:8080`      | HTTP listen address               |
| `DRIFT_DB`        | `drift.db`   | SQLite database file path         |
| `DRIFT_TMPL_DIR`  | (auto)       | Path to HTML template directory   |
| `DRIFT_STATIC_DIR`| (auto)       | Path to static assets (`web/static`) served at `/static/*` |

## Usage

1. **Upload price data** — go to `/data` and upload a CSV file (single-symbol `AAPL.csv`
   or multi-symbol with a `symbol` column).
2. **Create an experiment** — go to `/experiments/new`, select assets and weights,
   set simulation parameters, then save or run immediately.
3. **View results** — the Results page shows the percentile distribution chart and key
   statistics (p5/p25/p50/p75/p95, probability of loss, CAGR, max drawdown).

## CSV Format

Single-symbol:
```csv
date,open,high,low,close,volume,adjusted_close
2020-01-02,296.24,300.60,295.19,300.35,33389800,298.82
```

Multi-symbol:
```csv
date,symbol,open,high,low,close,volume,adjusted_close
2020-01-02,AAPL,296.24,300.60,295.19,300.35,33389800,298.82
2020-01-02,SPY,321.10,323.25,319.88,322.41,85234900,320.12
```

## Architecture

Hexagonal (ports & adapters), enforced by go-arch-lint:

```
adapters (http, ingestion, storage)  →  app  →  domain
```

- **Domain**: pure Go types in `internal/domain/` (assets, portfolios, experiments, results) — no external deps
- **App**: use-case orchestration (GBM engine, bootstrap resampling, CSV ingestion, results aggregation) in `internal/app/`, depending only on `domain` and `internal/ports/`
- **Ports**: interfaces the app exposes and requires, in `internal/ports/inbound/` and `internal/ports/outbound/`
- **HTTP adapter**: Chi router + Go `html/template` + HTMX in `internal/adapters/http/`
- **Storage adapter**: SQLite via `modernc.org/sqlite` (no CGo, no ORM) in `internal/adapters/storage/sqlite/`

See [`docs/architecture.md`](docs/architecture.md) for the full reference, including a component diagram.

## Simulation Models

| Model             | Description                                             |
|-------------------|---------------------------------------------------------|
| `gbm`             | Geometric Brownian Motion using historical μ and σ      |
| `bootstrap`       | Resample historical daily log-returns with replacement  |
| `block_bootstrap` | Same as bootstrap (block variant planned)               |
