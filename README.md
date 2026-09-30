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
updates since), and it isn't deployed on the homelab.

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
| `DRIFT_TMPL_DIR` | `internal/adapters/http/templates` in the source tree the binary was built from | HTML template directory |
| `DRIFT_STATIC_DIR` | `web/static` in the source tree the binary was built from | Static-asset directory for `/static/*` |

## How it works

Drift is a hexagonal (ports & adapters) Go application: an HTTP adapter serves a
server-rendered UI and calls into inbound ports, the domain/app core runs the
simulations against outbound ports, and a SQLite adapter persists assets,
experiments, and runs. Full reference: [docs/architecture.md](docs/architecture.md).

## Development

```bash
make check          # go fmt + go vet + golangci-lint + go test -race
go-arch-lint check  # hexagonal boundaries
go mod tidy         # CI fails if this changes go.mod or go.sum
```

CI runs on every push to `main` and every PR: `go build ./...`, `gofmt -l .`
(fails on any unformatted file rather than rewriting it), `go vet ./...`,
golangci-lint, `go test -race -count=1 -timeout 120s ./...`,
`go-arch-lint check`, and the `go mod tidy` diff check.
Conventions for contributors and agents: [AGENTS.md](AGENTS.md).

## Deployment

Pushing a `v*` tag runs [`release.yml`](.github/workflows/release.yml), which
builds standalone binaries (linux/darwin, amd64/arm64) and attaches them to a
GitHub release; no release has been tagged yet. There is no Dockerfile and no
homelab deployment. Templates and static assets are read from disk, not
embedded, so a binary run outside its source tree needs `DRIFT_TMPL_DIR` and
`DRIFT_STATIC_DIR` pointed at copies of `internal/adapters/http/templates` and
`web/static`, plus `DRIFT_ADDR` and `DRIFT_DB`.

## License

[Apache-2.0](LICENSE).
