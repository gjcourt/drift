<!-- readme-type: service -->
# Drift

Monte Carlo portfolio drift simulator

A single backtest or point forecast hides how wide the range of real outcomes for a
portfolio can be. Drift addresses that by running many simulated forward paths —
Geometric Brownian Motion or bootstrapped historical returns — over price data you
upload, then aggregates them into percentile bands, drawdown distributions,
probability of loss, and CAGR. It's for anyone who wants a distribution of outcomes
for a multi-asset portfolio, not a single number, from a self-hosted binary with no
external services.

**Status:** experimental — the simulation engine and web UI are feature-complete and
covered by CI, but there has been no feature work since 2026-07 (only dependency
updates since), and it isn't deployed anywhere yet.

## Quick start

Needs: Go 1.26 or newer.

```bash
git clone https://github.com/gjcourt/drift && cd drift
make dev
```

Then open http://localhost:8080.

## Usage

Upload price data on the `/data` page: a single-symbol CSV (`AAPL.csv`) or a
multi-symbol CSV with a `symbol` column. Formats are documented in
[docs/reference/2026-05-02-data-formats.md](docs/reference/2026-05-02-data-formats.md).

Create an experiment on `/experiments/new` — pick assets and weights, set the
simulation parameters (model, horizon, number of paths), then run it. The available
models are documented in
[docs/reference/2026-05-02-simulation-models.md](docs/reference/2026-05-02-simulation-models.md).

View a run's results page for the percentile chart and summary statistics
(p5/p25/p50/p75/p95, probability of loss, CAGR, max drawdown).

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `DRIFT_ADDR` | `:8080` | HTTP listen address |
| `DRIFT_DB` | `drift.db` | SQLite database file path |
| `DRIFT_TMPL_DIR` | (auto) | HTML template directory; auto-resolved from the source tree in dev |
| `DRIFT_STATIC_DIR` | (auto) | Static-asset directory for `/static/*` |

## How it works

Drift is a hexagonal (ports & adapters) Go application: an HTTP adapter serves a
server-rendered UI and calls into inbound ports, the domain/app core runs the
simulations against outbound ports, and a SQLite adapter persists assets,
experiments, and runs. Full reference: [docs/architecture.md](docs/architecture.md).

## Development

```bash
make fmt
make vet
make lint
make test
```

Or `make check` to run them all — this is what CI runs on every push and PR.
Conventions for contributors and agents: [AGENTS.md](AGENTS.md).

## Deployment

Drift ships as standalone binaries (linux/darwin, amd64/arm64) built by
[`release.yml`](.github/workflows/release.yml) on tagged releases — there is no
Dockerfile and no homelab deployment today. Self-host by running a binary with
`DRIFT_ADDR` and `DRIFT_DB` set (and `DRIFT_TMPL_DIR` / `DRIFT_STATIC_DIR` if the
templates and static assets aren't co-located with it).

## License

No licence file yet.
