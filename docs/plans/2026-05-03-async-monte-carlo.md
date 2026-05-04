---
title: "Async Monte Carlo execution"
status: "Draft"
created: "2026-05-03"
updated: "2026-05-03"
updated_by: "george"
tags: ["architecture", "simulation", "concurrency", "http"]
---

# Async Monte Carlo execution

## Problem statement

`POST /experiments/{id}/run` calls `SimulationService.RunExperiment` and
blocks the request goroutine for the entire Monte Carlo simulation
(`internal/adapters/http/handlers/simulations.go:10-18`,
`internal/app/simulation.go:27-56`). The handler then `303`s to
`/runs/{id}` only once the run has already completed.

Concrete failure modes this design has today:

1. **Client disconnect kills the run.** The simulator is invoked with
   `r.Context()`. If the user closes the tab, the upstream proxy times
   out, or the browser retries the form post, Chi cancels `r.Context()`,
   `assetRepo.GetPriceRecords` returns `context.Canceled`, and the run
   ends with `status="failed"`. Work is wasted; the user sees nothing.
2. **Reverse-proxy timeouts kill the run.** Defaults are 30–60s
   (Caddy 30s, nginx 60s, Cloudflare 100s). At `NumPaths=10000`,
   `HorizonDays=2520`, GBM runs ~30s on a laptop on cold caches; bigger
   configs reliably exceed any sane proxy default. The run keeps going
   in the worker pool but the HTTP response never lands and the user
   gets a 502/504.
3. **Status lifecycle is theatre.** The handler writes `status=running`
   then synchronously overwrites it with `complete`/`failed` on the
   same goroutine before the response is flushed. No reader ever
   observes `running`. The `runs.status` column documents a state
   machine the code does not actually traverse.
4. **No observability.** There is no way to see in-flight runs, their
   progress, how long they have been running, or to cancel one. Logs
   only show the request once it has finished.
5. **HTTP worker exhaustion.** `runtime.GOMAXPROCS` HTTP workers are
   blocked for the duration of the simulation. Three concurrent
   `RunExperiment` calls on a 4-core box ties up the whole server.

This plan moves simulation execution out of the request goroutine and
gives the run its own lifecycle, owned by the process — not the HTTP
request.

## Requirements

1. **Preserve the existing UI flow.** The experiment-detail page
   (`internal/adapters/http/templates/experiment-detail.html:13-15`)
   posts to `/experiments/{id}/run` and the user expects to land on
   `/runs/{id}`. That redirect must still work; the `/runs/{id}`
   page must render even when the run is still in progress.
2. **Allow long-running simulations.** A 50k-path × 10y-horizon GBM
   that takes 5+ minutes must complete reliably regardless of client
   or proxy behaviour.
3. **Show progress.** The results page must indicate `queued` →
   `running` → `complete`/`failed` without the operator having to
   refresh manually. Live percent-complete is desirable but not
   required for v1.
4. **Restart-safe.** A run that was in flight when the process died
   must either be resumed or be marked `failed` deterministically on
   restart — never left dangling at `running`.
5. **Idempotent retries.** Re-clicking "Run Simulation" must not
   silently double-execute. Same `experiment_id` + active run = error
   or queue-collapse.
6. **No new external services.** Drift is "a single Go binary plus an
   embedded SQLite file" (`AGENTS.md:77`). Solution must respect that.
7. **Honour hexagonal boundaries.** Worker code lives in
   `internal/app/`, schema additions in `internal/adapters/storage/sqlite/`,
   handler glue in `internal/adapters/http/handlers/`. The inbound port
   `inbound.SimulationService` is the seam.

## Architecture options

### Option A — In-memory worker pool + polling

Spawn a singleton `app.Runner` at process start with a buffered
`chan job`. The HTTP handler enqueues a job (`status=queued`),
returns 202 immediately, and a fixed pool of N goroutines pulls from
the channel and executes. The `/runs/{id}` page polls
`GET /runs/{id}` every second via a small JS fetch loop (or HTMX
`hx-trigger="every 1s"`).

| | |
|---|---|
| State model | Run row exists in SQLite for status; queue lives in Go memory |
| Restart behaviour | **Lossy.** Queued jobs vanish; `running` rows stay marked `running` forever |
| UI work | Trivial: meta-refresh or HTMX poll on the existing results page |
| Blast radius | Smallest; one new file in `internal/app/` |
| Concurrency control | Free — channel + N goroutines |
| Backpressure | Channel buffer; full = 503 |

### Option B — SQLite-backed job queue + worker goroutine + polling **(recommended)**

The `runs` table is the queue. Status column carries `queued` →
`running` → `complete` / `failed` / `cancelled`. A single
`app.Runner` goroutine (or a small fixed pool, see "Open questions")
polls every ~500ms for the oldest `queued` row, atomically claims it
(`UPDATE runs SET status='running', started_at=now WHERE id=? AND status='queued'`),
runs the simulation under a process-owned `context.Context`, and
writes the terminal state. On startup, a one-shot reaper marks any
rows still `running` from a previous process as `failed` with
`error="interrupted"`. UI polls `/runs/{id}` exactly as in Option A.

| | |
|---|---|
| State model | Single source of truth: SQLite `runs` table |
| Restart behaviour | **Safe.** Queued rows survive; reaper handles orphaned `running` rows; new runs queued during downtime get picked up |
| UI work | Same poll loop as A |
| Blast radius | Medium: schema migration + reaper + claim query + worker loop |
| Concurrency control | Configurable via `DRIFT_MAX_CONCURRENT_RUNS` (default: 1, since SQLite is single-writer) |
| Backpressure | Implicit: queue depth = count of `queued` rows; cap exposed via env var |

### Option C — SQLite-backed jobs + Server-Sent Events

Same persistence model as B, but the worker publishes per-run
progress (e.g. `paths_completed/total`) on an in-memory pub/sub
keyed by run ID, and `/runs/{id}/events` upgrades to SSE
(`Content-Type: text/event-stream`) and streams `progress`,
`complete`, and `failed` events. Falls back to polling for clients
that disconnect (the SSE is a *view*, not the system of record).

| | |
|---|---|
| State model | Same as B, plus an in-memory progress channel |
| Restart behaviour | Same as B |
| UI work | New SSE consumer in templates; needs a small JS handler. Live progress bar is the prize |
| Blast radius | B + SSE handler + progress reporting plumbed through the worker pool (`workerPool` needs a `progress chan int`) |
| Concurrency control | Same as B |
| Backpressure | Same as B |

### Option D — External queue (Redis / Postgres LISTEN)

Run state in Postgres or queue in Redis; worker is a separate process.

Rejected. It violates `AGENTS.md:77` ("Drift is self-contained: a
single Go binary plus an embedded SQLite file. No external
services."), trebles the operational surface area, and solves a
problem we don't have (multi-host scale-out) at the cost of one we
do (single-binary deploy).

## Recommendation

**Option B.** Implement Option C's SSE in a follow-up once B is
shipped and stable. Reasoning specific to this codebase:

- SQLite is already the system of record for runs
  (`internal/adapters/storage/sqlite/db.go:73-82`). Reusing the same
  table as the queue means zero new infra and one transactional
  source of truth.
- `db.SetMaxOpenConns(1)` (`db.go:29`) means we cannot meaningfully
  parallelise multiple writers anyway; a single worker goroutine is
  the natural fit and keeps the claim query trivially correct without
  needing `SKIP LOCKED` (which SQLite lacks).
- The hexagonal layout is already friendly to this:
  `inbound.SimulationService.RunExperiment` keeps the same signature
  but its semantics flip from "run synchronously" to "enqueue and
  return". The handler change is a one-liner.
- Polling at 1Hz on `/runs/{id}` is cheap (one indexed point read)
  and the existing template renders all three states already
  (`results.html:7,57`). SSE would be nicer UX but adds a JS
  consumer, a pub/sub, and a streaming handler — all worth doing
  *after* the request-path bug is fixed, not as part of fixing it.
- Restart-safety is a stated requirement; Option A cannot meet it.

## Schema additions

Migration applied alongside the existing `schema` constant in
`internal/adapters/storage/sqlite/db.go:42`. New columns are added
idempotently via `ALTER TABLE … ADD COLUMN` guarded by a `PRAGMA
table_info` check (or, simpler, with `IF NOT EXISTS` semantics
emulated by inspecting `sqlite_master`).

```sql
-- Existing runs table grows three columns:
ALTER TABLE runs ADD COLUMN queued_at      TEXT NOT NULL DEFAULT '';
ALTER TABLE runs ADD COLUMN claimed_at     TEXT;          -- nullable
ALTER TABLE runs ADD COLUMN progress_total INTEGER NOT NULL DEFAULT 0;
ALTER TABLE runs ADD COLUMN progress_done  INTEGER NOT NULL DEFAULT 0;

-- Index for the claim query and the experiment-detail listing.
CREATE INDEX IF NOT EXISTS idx_runs_status_queued_at ON runs(status, queued_at);
CREATE INDEX IF NOT EXISTS idx_runs_experiment_id    ON runs(experiment_id, started_at DESC);
```

Status state machine (`runs.status`):

```
            +---------+
  enqueue → | queued  | ─── claim ────┐
            +---------+               ▼
                                 +----------+
                                 | running  | ─── compute_stats ──► +-----------+
                                 +----------+                       | complete  |
                                  │  │                              +-----------+
                                  │  └── error ──► +--------+
                                  │                | failed |
                                  │                +--------+
                                  └── shutdown_reap ► failed (error="interrupted")
```

`StatusQueued` is added to `internal/domain/experiment.go` alongside
the existing four:

```go
const (
    StatusDraft     ExperimentStatus = "draft"
    StatusQueued    ExperimentStatus = "queued"     // NEW
    StatusRunning   ExperimentStatus = "running"
    StatusComplete  ExperimentStatus = "complete"
    StatusFailed    ExperimentStatus = "failed"
)
```

The `domain.Run` struct grows two fields to surface progress to the
results page (zero values render fine; existing tests are
unaffected):

```go
type Run struct {
    ID           string
    ExperimentID string
    QueuedAt     time.Time   // NEW
    StartedAt    time.Time   // becomes the "claimed_at"; renamed comment only
    FinishedAt   *time.Time
    Status       ExperimentStatus
    Error        string
    Stats        ResultStats
    ProgressDone  int        // NEW; 0 until worker reports
    ProgressTotal int        // NEW; equals NumPaths once running
}
```

## API changes

| Method | Path | Before | After |
|---|---|---|---|
| `POST` | `/experiments/{id}/run` | 200/303 after run completes | **202 Accepted** with `Location: /runs/{id}` and HTML body containing the same redirect (so the existing form post still ends up on the run page) |
| `GET`  | `/runs/{id}` | Renders complete-or-failed page | Renders queued/running/complete/failed; auto-refreshes (`<meta http-equiv="refresh" content="2">` on non-terminal states, or HTMX `hx-trigger="every 2s"` on a partial) |
| `GET`  | `/runs/{id}.json` | _new_ | Returns `{id, status, progress_done, progress_total, error?, stats?}` for poll consumers (small JSON; cheap; cache-control: no-store) |
| `POST` | `/runs/{id}/cancel` | _new (optional, see Open questions)_ | Sets `status=cancelled` if currently `queued` or `running`; worker observes via shared cancel context keyed by run ID |

The form-post HTML flow does **not** need changing in v1: the browser
treats 202 + `Location` like a 303 in practice as long as we either
return 303 (we will, to keep behaviour identical for the form
submitter) or 202 with an HTML body that contains a `<meta refresh>`
to the run page. Recommendation: **return 303 to `/runs/{id}` exactly
as today** — the only behavioural change is that the run hasn't
finished yet when the redirect fires.

Handler diff (sketch):

```go
// internal/adapters/http/handlers/simulations.go
func (h *H) RunExperiment(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    run, err := h.sim.EnqueueRun(r.Context(), id) // was: RunExperiment
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/runs/"+run.ID, http.StatusSeeOther)
}
```

`inbound.SimulationService` gains one method and the existing one
keeps its name but its semantics shift to "enqueue":

```go
type SimulationService interface {
    EnqueueRun(ctx context.Context, experimentID string) (*domain.Run, error)
    GetRun(ctx context.Context, runID string) (*domain.Run, error)
    CancelRun(ctx context.Context, runID string) error // optional v1.1
    // GetRunPaths is removed (it's a stub today; see critique Moderate)
}
```

`EnqueueRun` writes the row with `status=queued`, `queued_at=now`,
`progress_total=cfg.NumPaths`, and returns. The `ctx` it receives is
`r.Context()` — fine, because nothing long-running happens here.

## Worker implementation sketch

New file: `internal/app/runner.go`.

```go
package app

// Runner owns the lifecycle of background simulation execution.
// It is constructed once at process start and shut down on SIGTERM.
type Runner struct {
    sim    *simulationSvc          // private — Runner reuses the simulate() core
    repo   outbound.SimulationRepository
    expr   outbound.ExperimentRepository
    logger *slog.Logger

    pollInterval time.Duration     // default 500ms
    concurrency  int               // default 1; configurable

    cancels sync.Map               // runID → context.CancelFunc for /cancel
}

// NewRunner wires a Runner. It does not start it; call Start.
func NewRunner(sim *simulationSvc, repo outbound.SimulationRepository,
    expr outbound.ExperimentRepository, opts ...RunnerOption) *Runner { ... }

// Start launches the worker loop. It returns when ctx is done and all
// in-flight runs have finished or been cancelled.
func (r *Runner) Start(ctx context.Context) error {
    if err := r.reapOrphans(ctx); err != nil { return err } // mark stale 'running' as failed
    sem := make(chan struct{}, r.concurrency)
    for {
        select {
        case <-ctx.Done():
            // wait for in-flight; sem fills back up when goroutines exit
            for i := 0; i < r.concurrency; i++ { sem <- struct{}{} }
            return nil
        default:
        }
        run, ok, err := r.claimNext(ctx)
        if err != nil { r.logger.Error("claim", "err", err); time.Sleep(r.pollInterval); continue }
        if !ok { time.Sleep(r.pollInterval); continue }
        sem <- struct{}{}
        go func(run domain.Run) {
            defer func() { <-sem }()
            r.execute(ctx, run)
        }(run)
    }
}

// claimNext atomically transitions one queued row to running and returns it.
// The UPDATE … WHERE status='queued' is race-safe because SQLite serialises
// writes (SetMaxOpenConns(1) in storage/sqlite).
func (r *Runner) claimNext(ctx context.Context) (domain.Run, bool, error) {
    // SELECT id FROM runs WHERE status='queued' ORDER BY queued_at ASC LIMIT 1;
    // UPDATE runs SET status='running', started_at=? WHERE id=? AND status='queued';
    // (rows affected == 1) → load the row and return
}

// reapOrphans is called once on Start. Any row still 'running' from a
// previous process is marked 'failed' with error="interrupted".
func (r *Runner) reapOrphans(ctx context.Context) error {
    // UPDATE runs SET status='failed', error='interrupted', finished_at=?
    //   WHERE status='running';
}

// execute runs the simulation under a context the Runner owns
// (NOT r.Context()). It writes terminal state on completion/failure.
func (r *Runner) execute(parent context.Context, run domain.Run) { ... }
```

Wiring in `cmd/drift/main.go`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

simSvc := app.NewSimulationService(store, store, store)
runner := app.NewRunner(simSvc, store, store,
    app.WithConcurrency(envIntOr("DRIFT_MAX_CONCURRENT_RUNS", 1)),
)
go func() {
    if err := runner.Start(ctx); err != nil { slog.Error("runner", "err", err) }
}()

srv := &http.Server{
    Addr:              addr,
    Handler:           handler,
    ReadHeaderTimeout: 10 * time.Second,
    IdleTimeout:       60 * time.Second,
}
go func() { _ = srv.ListenAndServe() }()

<-ctx.Done()
shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
_ = srv.Shutdown(shutCtx) // stop accepting requests first
runner.Wait()             // then drain in-flight runs (or forcibly cancel after timeout)
```

The `simulationSvc` keeps its existing `simulate`/`runGBM`/`runBootstrap`
methods. They are now called only from `Runner.execute`, never
directly from the HTTP handler. `EnqueueRun` is a thin new method on
`simulationSvc` that just writes a `queued` row.

Graceful shutdown ordering matters: HTTP server first (no new
enqueues), then `runner.Wait()` (drain in-flight). On SIGTERM-with-
running-job, the Runner respects an outer timeout (`DRIFT_RUN_DRAIN_TIMEOUT`,
default 30s); past that, it cancels the run's context, the worker
records `failed` with `error="interrupted"`, and exits. The reaper
on next start will not see it (already terminal).

## Migration path

The contract change to `inbound.SimulationService.RunExperiment` →
`EnqueueRun` is the only breaking part. Since this is internal and
the only caller is the HTTP handler in this same repo, it's a
mechanical refactor in one PR.

For tests that want the old synchronous behaviour, **do not** add a
query-param escape hatch — that's a path to keeping two code paths
forever. Instead, expose a test-only `Runner.RunInline(ctx, runID)`
helper that the HTTP integration test calls directly after
`EnqueueRun` to drain the queue synchronously. This keeps the
production path single and the test path explicit. (See "Test
plan".)

Existing data is untouched: the migration adds nullable / default-
valued columns. Old `runs` rows continue to render fine.

## Test plan

### Unit — `internal/app/runner_test.go` (new)

- `TestRunner_ClaimsAndCompletes`: enqueue 3 runs, start runner with
  concurrency=1, assert all three reach `complete` in order, stats
  populated.
- `TestRunner_FailedRunIsRecorded`: inject a `simulationRepo` whose
  `GetExperiment` returns an error; enqueue; assert `status=failed`,
  `error` non-empty, `finished_at` set.
- `TestRunner_ContextCancelDuringRun`: start a long synthetic run
  (config with `NumPaths=1_000_000`), cancel the runner context,
  assert run ends as `failed` with `error="interrupted"`, no
  `running` rows left.
- `TestRunner_ReapOrphansOnStart`: pre-seed a `runs` row with
  `status=running`; start runner; assert it transitions to `failed`
  before claiming any new work.
- `TestRunner_ClaimIsAtomic`: spin up two `Runner` instances against
  the same store (sanity, even though prod is single-process); enqueue
  one row; assert exactly one runner claims it.

### Unit — `internal/adapters/storage/sqlite/db_test.go` (extend)

- `TestSaveRun_PersistsQueuedAtAndProgress`.
- `TestClaimNextQueued_SkipsRunningAndComplete`.
- `TestReapStaleRunning_OnlyTouchesRunning`.

### Integration — `internal/adapters/http/handlers/simulations_test.go` (new)

Uses `httptest.NewServer` against a real SQLite tempfile and a
`Runner` with `concurrency=1` started in the test:

- `TestRunExperiment_EnqueuesAndCompletes`: POST run → 303 →
  poll `/runs/{id}.json` until `status=complete` or 5s timeout →
  assert stats.
- `TestRunExperiment_ClientDisconnectDoesNotKillRun`: POST run with a
  `r.Context()` cancelled immediately after the redirect; assert the
  run still completes (this is **the** regression test for the bug
  this plan fixes).
- `TestGetRunJSON_QueuedRunReturnsProgressZero`.

### Manual smoke

- Start `drift` locally; create a 50k-path × 10y-horizon experiment
  via the UI; verify the page returns instantly, the results page
  shows `running`, then transitions to `complete` without browser
  intervention.
- `kill -TERM` the process while a run is in flight; restart; verify
  the run is now `failed` with `error="interrupted"` (not stuck at
  `running`).

## Open questions

1. **Run retention.** Should completed/failed runs be auto-pruned
   after N days? The `runs.stats` JSON is small, but unbounded growth
   over years still bloats the DB. Suggest a `DRIFT_RUN_RETENTION_DAYS`
   env var, default 0 (disabled), with a daily janitor goroutine.
2. **Max concurrent runs.** SQLite is single-writer, but the
   simulation itself doesn't write between start and end — it reads
   prices once, computes in memory, and writes one row at the end.
   We could safely run 2–4 concurrently on a multicore box. Default
   1 for safety; expose `DRIFT_MAX_CONCURRENT_RUNS`. Question: should
   per-run worker-pool size (`runtime.NumCPU()` today) shrink when
   concurrency > 1, to avoid CPU thrashing?
3. **Surface partial / progress?** v1 plan stores
   `progress_done`/`progress_total` but the worker only updates
   them at start (0/N) and end (N/N). To get live progress we need
   to plumb a `progress chan int` through `workerPool` in
   `simulation.go:189`. Cheap to add (one buffered channel, a
   batched `UPDATE runs SET progress_done=? WHERE id=?` every ~250ms),
   but it's a separate PR. Worth doing once SSE lands (Option C).
4. **Cancel UX.** Should the run-detail page expose a "Cancel"
   button? Trivial server-side (close the registered cancel func);
   needs a confirm dialog and one new POST handler. Optional v1.1.
5. **SSE follow-up.** Once B is shipped, what's the backpressure
   strategy for slow SSE clients? Drop oldest progress event, or
   block? Suggest: per-client ring buffer of size 16, drop oldest
   on overflow.
6. **Reaper aggressiveness.** Should the reaper distinguish "process
   crashed mid-run" (mark failed) from "another instance is running
   this" (no-op)? In a strict single-process deploy this doesn't
   matter; if we ever run two `drift` binaries against the same DB
   we need a heartbeat column (`runs.heartbeat_at`) and a staleness
   threshold. Out of scope for v1.
