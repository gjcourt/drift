# Architecture

Drift is structured around **hexagonal architecture** (ports and adapters), also known as the "clean architecture" style. The goal is to keep business logic isolated from infrastructure concerns so each layer can be tested and replaced independently.

---

## Package Map

```
github.com/gjcourt/drift
├── cmd/drift/              ← Composition root: wires all dependencies and starts HTTP server
└── internal/
    ├── domain/             ← Core business types and pure logic (no external imports)
    ├── ports/
    │   ├── inbound/        ← Interfaces that services expose to the outside world (HTTP handlers)
    │   └── outbound/       ← Interfaces that services consume from infrastructure (DB, storage)
    ├── services/           ← Application logic; implements inbound ports, depends on outbound ports
    └── adapters/
        ├── http/           ← Chi router, HTML/template handlers, static files
        ├── ingestion/      ← CSV and JSON parsers
        └── storage/
            └── sqlite/     ← SQLite implementation of all outbound repository ports
```

---

## Dependency Rule

Dependencies flow **inward only**:

```
adapters  →  services  →  domain
              ↕ (via ports)
        outbound ports  →  domain
```

- `domain` imports nothing from this module.
- `ports` import only `domain`.
- `services` import `domain` and `ports` only.
- `adapters` import `ports`, `domain`, and external libraries.
- `cmd/drift` imports everything and wires it together.

This means you can swap any adapter (e.g. replace SQLite with Postgres) without touching the domain or service layers.

---

## Hexagonal Diagram

```
       ┌─────────────────────────────────────────────┐
       │                 cmd/drift                   │
       │           (composition root)                │
       └───────────┬────────────────────┬────────────┘
                   │                    │
       ┌───────────▼──────┐    ┌────────▼────────────┐
       │  adapters/http   │    │  adapters/storage/  │
       │  (Chi + templates│    │      sqlite         │
       │   HTMX frontend) │    │  (SQL, go-sqlite3)  │
       └───────────┬──────┘    └────────┬────────────┘
                   │                    │
       ┌───────────▼────────────────────▼────────────┐
       │                  ports/                     │
       │    inbound: DataIngestionService             │
       │             ResultsService                  │
       │             SimulationService               │
       │   outbound: AssetRepository                 │
       │             ExperimentRepository            │
       │             SimulationRepository            │
       └───────────────────┬─────────────────────────┘
                           │
       ┌───────────────────▼─────────────────────────┐
       │                services/                    │
       │   ingestionSvc · resultsSvc · simulationSvc │
       └───────────────────┬─────────────────────────┘
                           │
       ┌───────────────────▼─────────────────────────┐
       │                  domain/                    │
       │  Asset · Experiment · SimulationConfig      │
       │  SimulatedPath · ResultStats · Portfolio    │
       └─────────────────────────────────────────────┘
```

---

## HTTP Layer

The HTTP adapter uses [Chi](https://github.com/go-chi/chi) as the router and Go's `html/template` for server-side rendering, augmented with [HTMX](https://htmx.org) for partial-page updates without a separate JS framework.

### Template rendering

All page templates share a single `layout.html` base. To avoid the Go `html/template` shared-namespace problem (where `{{define "content"}}` blocks across multiple files collide when parsed together), the adapter uses a **clone-per-request** pattern:

1. At startup, `layout.html` is parsed once into a `*template.Template` called `baseTmpl`.
2. Each request handler calls `h.page("foo.html")`, which clones `baseTmpl` and then parses only the requested page template into the clone.
3. The handler renders via `ExecuteTemplate(w, "layout", data)`.

This trades a small per-request allocation for correct template isolation.

### Route table

| Method | Path                    | Handler               | Description                            |
|--------|-------------------------|-----------------------|----------------------------------------|
| GET    | `/`                     | `Dashboard`           | Summary of assets & experiments        |
| GET    | `/data`                 | `DataManager`         | List uploaded assets                   |
| POST   | `/data/upload`          | `UploadCSV`           | Upload a CSV file of price history     |
| DELETE | `/data/{symbol}`        | `DeleteAsset`         | Remove an asset and its price records  |
| GET    | `/experiments`          | `ListExperiments`     | Experiment index                       |
| GET    | `/experiments/new`      | `NewExperimentForm`   | Experiment builder form                |
| POST   | `/experiments`          | `CreateExperiment`    | Submit a new experiment                |
| GET    | `/experiments/{id}`     | `ExperimentDetail`    | Experiment detail & run history        |
| POST   | `/experiments/{id}/run` | `RunExperiment`       | Trigger a synchronous simulation run   |
| GET    | `/runs/{id}`            | `RunResults`          | Simulation results page                |
| GET    | `/static/*`             | `FileServer`          | Static assets (CSS, JS)                |

---

## Storage

SQLite is used via [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite), a pure-Go CGo-free driver. The `Store` type in `internal/adapters/storage/sqlite` implements all three outbound repository interfaces from a single database connection.

Schema is applied at startup via `CREATE TABLE IF NOT EXISTS` statements in `New()`.

---

## Concurrency model

Monte Carlo simulation runs on a **worker pool** sized to `runtime.NumCPU()`. Each worker gets its own `rand.Rand` seeded independently from a base seed (either from `SimulationConfig.Seed` for reproducibility, or `time.Now().UnixNano()` for non-deterministic runs). Results are collected over a buffered channel and assembled into a `[]SimulatedPath` before stats are computed.
