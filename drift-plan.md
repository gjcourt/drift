# Drift — Monte Carlo Stock Simulation Engine

> **Drift** takes its name from the *drift* parameter ($\mu$) in Geometric Brownian Motion — the expected rate of return that pulls a price path forward through time while randomness governs the journey.

---

## 1. Project Overview

Drift is a self-hosted Monte Carlo simulation engine for exploring alternative historical and forward-looking paths for stock portfolios. Users upload historical price data, configure simulation parameters, stage experiments, run them, and explore results through an interactive web UI.

**Core capabilities:**
- Ingest historical OHLCV data via CSV or JSON
- Define multi-asset portfolios with weighted allocations
- Run Monte Carlo simulations using Geometric Brownian Motion (GBM) or bootstrapped-returns models
- Visualise fan charts, percentile bands, drawdown distributions, and summary statistics
- Store and compare multiple simulation runs (experiments)

---

## 2. Catchy Name: **Drift**

Selected from the following shortlist:

| Name | Rationale |
|---|---|
| **Drift** ⭐ | The GBM drift term $\mu$; evokes wandering, uncertainty, time |
| Wander | Random walk; memorable but generic |
| Brownian | Scientifically precise; less accessible |
| Pathforge | Action-oriented; slightly long |
| Parallax | Different perspectives on the same future |

**Drift** wins: short, memorable, scientifically grounded, and domain-evocative.

---

## 3. Architecture — Hexagonal (Ports & Adapters)

```
┌──────────────────────────────────────────────────────────┐
│                      Driving Side                         │
│   HTTP Adapter (REST + Server-Side HTML)                  │
│   CLI Adapter (optional, for scripted runs)               │
└─────────────────────┬────────────────────────────────────┘
                      │  Inbound Ports
┌─────────────────────▼────────────────────────────────────┐
│                     Domain Core                           │
│                                                          │
│   SimulationService    DataIngestionService               │
│   ExperimentService    ResultsService                     │
│                                                          │
│   Domain Models:                                         │
│     Asset · Portfolio · Simulation · Path · Result       │
│     Experiment · PriceRecord · Statistics                │
└─────────────────────┬────────────────────────────────────┘
                      │  Outbound Ports
┌─────────────────────▼────────────────────────────────────┐
│                     Driven Side                           │
│   SQLite Adapter (asset + experiment persistence)         │
│   File Adapter (CSV / JSON ingestion)                     │
│   In-Memory Adapter (ephemeral test/dev)                  │
└──────────────────────────────────────────────────────────┘
```

### 3.1 Directory Structure

```
drift/
├── cmd/
│   └── drift/
│       └── main.go              # Entry point: wire adapters → start server
├── internal/
│   ├── domain/                  # Pure domain types — no imports from outside domain
│   │   ├── asset.go             # Asset, PriceRecord value objects
│   │   ├── portfolio.go         # Portfolio aggregate
│   │   ├── simulation.go        # Simulation config + parameters
│   │   ├── path.go              # SimulatedPath value object
│   │   ├── result.go            # SimulationResult aggregate
│   │   └── experiment.go        # Experiment (named collection of simulation runs)
│   ├── ports/
│   │   ├── inbound/
│   │   │   ├── simulation.go    # SimulationService interface
│   │   │   ├── ingestion.go     # DataIngestionService interface
│   │   │   └── results.go       # ResultsService interface
│   │   └── outbound/
│   │       ├── asset_repo.go    # AssetRepository interface
│   │       ├── simulation_repo.go
│   │       └── experiment_repo.go
│   ├── services/                # Domain service implementations
│   │   ├── simulation.go        # Monte Carlo engine logic
│   │   ├── ingestion.go         # Data ingestion + validation
│   │   └── results.go           # Result aggregation + statistics
│   └── adapters/
│       ├── http/
│       │   ├── server.go        # Chi/stdlib HTTP router setup
│       │   ├── handlers/
│       │   │   ├── assets.go
│       │   │   ├── experiments.go
│       │   │   └── simulations.go
│       │   └── templates/       # Go html/template files
│       │       ├── layout.html
│       │       ├── dashboard.html
│       │       ├── data-manager.html
│       │       ├── experiment-builder.html
│       │       ├── run.html
│       │       └── results.html
│       ├── storage/
│       │   └── sqlite/          # SQLite + sqlc generated code
│       │       ├── queries.sql
│       │       ├── schema.sql
│       │       └── db.go
│       └── ingestion/
│           ├── csv.go           # CSV file parser adapter
│           └── json.go          # JSON config parser adapter
├── web/
│   └── static/
│       ├── drift.css            # Minimal custom styles
│       └── drift.js             # Chart rendering (Chart.js via CDN)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 4. Data Formats

### 4.1 Historical Price Data (CSV ingestion)

Primary ingest format. One file per symbol, or multi-symbol in one file.

**Single-symbol file** (`AAPL.csv`):
```csv
date,open,high,low,close,volume,adjusted_close
2020-01-02,296.24,300.60,295.19,300.35,33389800,298.82
2020-01-03,297.15,300.58,296.50,297.43,29601200,295.93
2020-01-06,293.79,299.96,292.75,299.80,29596300,298.22
```

**Multi-symbol file** (`prices.csv`):
```csv
date,symbol,open,high,low,close,volume,adjusted_close
2020-01-02,AAPL,296.24,300.60,295.19,300.35,33389800,298.82
2020-01-02,SPY,321.10,323.25,319.88,322.41,85234900,320.12
2020-01-03,AAPL,297.15,300.58,296.50,297.43,29601200,295.93
```

**Field rules:**
- `date`: ISO 8601 (`YYYY-MM-DD`), trading days only
- `adjusted_close`: required for return calculations (handles splits/dividends)
- All price fields: decimal, USD-denominated
- `volume`: integer, may be omitted (set to `0`)
- Rows with missing `adjusted_close` are rejected with a warning; others are skipped silently

### 4.2 Simulation Configuration (JSON)

Used for staging experiments programmatically or as the UI's persistence format.

```json
{
  "version": "1",
  "experiment": {
    "name": "Retirement Portfolio — 10yr Outlook",
    "description": "Conservative 70/30 mix, 252-day trading year horizon"
  },
  "portfolio": {
    "assets": [
      { "symbol": "SPY",  "weight": 0.70 },
      { "symbol": "BND",  "weight": 0.30 }
    ],
    "rebalance": "annual"
  },
  "simulation": {
    "model": "gbm",
    "num_paths": 1000,
    "horizon_days": 2520,
    "lookback_days": 756,
    "start_value": 100000.00,
    "seed": null
  },
  "parameters": {
    "annual_contribution": 12000.00,
    "withdrawal_rate": null
  }
}
```

**Model options:**
| `model` | Description |
|---|---|
| `gbm` | Geometric Brownian Motion — uses historical $\mu$ and $\sigma$ |
| `bootstrap` | Resample historical daily returns with replacement |
| `block_bootstrap` | Block resample to preserve autocorrelation |

### 4.3 Results Export (JSON)

Persisted after each run; also downloadable from the UI.

```json
{
  "experiment_id": "exp_01JN4KZ8M3",
  "run_id": "run_01JN4KZA91",
  "ran_at": "2026-02-27T14:32:01Z",
  "num_paths": 1000,
  "percentiles": {
    "p5":  62340.12,
    "p25": 89201.44,
    "p50": 118432.91,
    "p75": 156009.32,
    "p95": 234881.55
  },
  "probability_of_loss": 0.087,
  "max_drawdown": {
    "median": -0.182,
    "p95":    -0.441
  },
  "paths": [
    [100000, 101240, 99830, "..."],
    "..."
  ]
}
```

> `paths` is omitted from the summary view and only returned when explicitly requested (large payload).

---

## 5. Web UI Design

Stack: **Go `html/template` + HTMX + Chart.js**. No separate frontend build step. All pages server-rendered; HTMX handles partial updates for a SPA feel without a JS framework.

### 5.1 Page Map

```
/ ─────────────── Dashboard
/data ─────────── Data Manager
/data/upload ──── Upload historical prices
/experiments ──── Experiment list
/experiments/new ─ Experiment Builder (stage a simulation)
/experiments/:id ─ Experiment detail + run history
/runs/:id ──────── Results Viewer
```

### 5.2 Page Descriptions

#### Dashboard (`/`)
- Summary cards: total assets loaded, experiments staged, simulations run
- Recent runs table with quick-link to results
- "New Experiment" CTA button

#### Data Manager (`/data`)
- Table of all ingested assets: symbol, date range, number of records, last updated
- Upload form: drag-and-drop or file picker for CSV; auto-detects single/multi-symbol
- Per-symbol: sparkline of adjusted close history; delete button

#### Experiment Builder (`/experiments/new`)
Multi-step form (HTMX-driven, single page):
1. **Name & Description** — free text
2. **Asset Selection** — search loaded symbols, add to portfolio, set weights (must sum to 100%)
3. **Simulation Parameters** — model selector, num paths slider (100–10 000), horizon (days), lookback window, starting value, annual contribution/withdrawal
4. **Review & Stage** — read-only summary; "Save as Draft" or "Save & Run Now"

#### Experiment Detail (`/experiments/:id`)
- Experiment config summary (read-only; "Clone to new" button)
- Run history table: run ID, timestamp, num paths, p50 terminal value, status (running / complete / failed)
- "Run Again" button (re-runs with same config)
- Real-time progress via HTMX polling during an active run

#### Results Viewer (`/runs/:id`)
- **Fan chart**: all simulated paths rendered as translucent lines with bold p5/p25/p50/p75/p95 percentile bands overlaid — rendered with Chart.js
- **Terminal value histogram**: distribution of final portfolio values
- **Drawdown chart**: distribution of maximum drawdown per path
- **Summary statistics table**: mean, median, std dev, Sharpe estimate, probability of loss, CAGR at each percentile
- Export buttons: download results JSON, download paths CSV

### 5.3 UI Wireframe — Results Viewer

```
┌──────────────────────────────────────────────────────────┐
│ Drift           Dashboard  Data  Experiments             │
├──────────────────────────────────────────────────────────┤
│ Run: run_01JN4KZA91   Experiment: Retirement 10yr        │
│ 1 000 paths · GBM · 2 520 days · Started $100 000       │
├──────────┬───────────────────────────────────────────────┤
│ p95      │                    ╱ ─ ─ ─ ─ ─ ─             │
│ $234 881 │                  ╱  ░░░░░░░░░░░░░            │
│          │               ╱    ░░░░░░░░░░░░░            │
│ p50      │            ╱       ▓▓▓▓▓▓▓▓▓▓▓▓            │
│ $118 432 │         ╱          ▓▓▓▓▓▓▓▓▓▓▓▓            │
│          │      ╱             ░░░░░░░░░░░░░            │
│ p5       │   ╱                ░░░░░░░░░░░░░            │
│ $62 340  │╱ ─ ─ ─ ─ ─                                  │
│          └───────────────────────────────────────────── │
│          Year 0          Year 5           Year 10       │
├──────────┴───────────────────────────────────────────────┤
│ Probability of loss: 8.7%    Median CAGR: 6.4%          │
│ Median max drawdown: -18.2%  p95 drawdown: -44.1%       │
└──────────────────────────────────────────────────────────┘
```

---

## 6. Simulation Engine — Key Design Decisions

### 6.1 Geometric Brownian Motion

Each asset $i$ follows:
$$S_t^{(i)} = S_0^{(i)} \cdot \exp\!\left(\left(\mu_i - \tfrac{1}{2}\sigma_i^2\right)t + \sigma_i W_t\right)$$

Where $\mu_i$ and $\sigma_i$ are estimated from the historical `adjusted_close` log-returns over the lookback window, and $W_t$ is a standard Brownian motion.

For multi-asset portfolios, Drift samples from a **multivariate normal distribution** using the Cholesky decomposition of the historical correlation matrix — ensuring realistic co-movement between assets.

### 6.2 Concurrency Model

Each simulation run spawns a worker pool (bounded by `runtime.NumCPU()`). Each worker generates a batch of paths independently. Results are streamed into a channel and aggregated once all workers complete.

```
RunSimulation()
  └─ for each batch → go worker(seed, batchSize, params, resultCh)
                         └─ generates paths
                         └─ sends []Path to resultCh
  └─ aggregator goroutine drains resultCh → builds Result
```

### 6.3 Reproducibility

Passing a non-nil `seed` in the simulation config pins the PRNG (using `math/rand/v2` with a `ChaCha8` source) so runs are fully reproducible.

---

## 7. Tech Stack

| Layer | Choice | Rationale |
|---|---|---|
| Language | Go 1.23+ | Performance, concurrency, single binary |
| HTTP router | `net/http` + `chi` | Lightweight, idiomatic |
| Templating | `html/template` | No separate build; safe by default |
| Partial updates | HTMX | SPA feel without a JS framework |
| Charts | Chart.js (CDN) | Minimal JS, well-documented |
| Database | SQLite via `modernc.org/sqlite` | Zero-dependency, file-based, sufficient for personal use |
| Query generation | `sqlc` | Type-safe SQL; avoids ORM overhead |
| Config | env vars + optional `config.yaml` | 12-factor friendly |
| Build | `Makefile` + `go build` | Single binary output |
| Containerisation | `Dockerfile` (distroless) | Easy self-hosting |

---

## 8. Milestones

| # | Milestone | Deliverables |
|---|---|---|
| 1 | **Core domain + CSV ingestion** | Domain types, CSV adapter, in-memory repo, unit tests |
| 2 | **GBM simulation engine** | Single-asset GBM, worker pool, reproducible paths, stats aggregation |
| 3 | **SQLite persistence** | Schema, sqlc queries, asset + experiment repos |
| 4 | **HTTP API + server-rendered UI** | All routes, Go templates, HTMX wiring |
| 5 | **Multi-asset + correlation** | Cholesky sampling, portfolio rebalancing |
| 6 | **Results visualisation** | Chart.js fan chart, histogram, drawdown chart |
| 7 | **Bootstrap model** | Daily-return resampling as alternative to GBM |
| 8 | **Containerisation + docs** | Dockerfile, docker-compose, README |

---

## 9. Open Questions

1. **Data sourcing**: Should Drift include a built-in provider adapter (e.g., Yahoo Finance unofficial API) to fetch historical data automatically, or remain import-only?
2. **Authentication**: Single-user (no auth) or lightweight user accounts? Relevant for self-hosted vs. shared deployments.
3. **Contribution modelling**: Should periodic contributions/withdrawals be modelled as deterministic cash flows injected at each rebalance, or stochastically?
4. **Stress scenarios**: Should users be able to pin a segment of historical returns (e.g., inject a 2008-style crash) into paths?

---

*Plan authored: 2026-02-27*
