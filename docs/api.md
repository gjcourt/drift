# API Reference

Drift is a **server-rendered web application**. There is no REST/JSON API;
every route returns HTML rendered by Go `html/template`. The frontend uses
[HTMX](https://htmx.org/) to swap fragments without full-page reloads.

---

## Route Table

| Method   | Path                    | Handler               | Description                              |
|----------|-------------------------|-----------------------|------------------------------------------|
| `GET`    | `/`                     | `Dashboard`           | Landing page — portfolio summary         |
| `GET`    | `/data`                 | `DataManager`         | Manage uploaded price data               |
| `POST`   | `/data/upload`          | `UploadCSV`           | Upload a CSV price file                  |
| `DELETE` | `/data/{symbol}`        | `DeleteAsset`         | Remove all price data for a symbol       |
| `GET`    | `/experiments`          | `ListExperiments`     | List all experiments                     |
| `GET`    | `/experiments/new`      | `NewExperimentForm`   | Render the new-experiment form           |
| `POST`   | `/experiments`          | `CreateExperiment`    | Create (and optionally run) an experiment|
| `GET`    | `/experiments/{id}`     | `ExperimentDetail`    | View a single experiment's details       |
| `POST`   | `/experiments/{id}/run` | `RunExperiment`       | Trigger a simulation run                 |
| `GET`    | `/runs/{id}`            | `RunResults`          | View simulation results for a run        |
| `GET`    | `/static/*`             | `http.FileServer`     | Static assets (JS, CSS, vendor libs)     |

### Path Parameters

| Parameter | Type   | Description                       |
|-----------|--------|-----------------------------------|
| `{symbol}`| string | Ticker symbol (e.g. `AAPL`)       |
| `{id}`    | int64  | Experiment or run ID (row key)    |

---

## Endpoint Details

### `GET /`

Renders the dashboard with a portfolio overview (symbols loaded, experiment
count). No request parameters.

---

### `GET /data`

Renders a table of all uploaded symbols with their record counts and date
ranges, plus an upload form.

---

### `POST /data/upload`

Upload a CSV file containing historical price data.

**Request**: `multipart/form-data`

| Field  | Type | Required | Description                                         |
|--------|------|----------|-----------------------------------------------------|
| `file` | file | yes      | The `.csv` file. Filename determines the symbol when no `symbol` column is present (e.g. `AAPL.csv` → symbol `AAPL`). Max 32 MiB. |

**Success response**: HTTP 200 with `HX-Redirect: /data` header. HTMX
intercepts this header and performs a client-side redirect to `/data`.

**Error response**: HTTP 400 or 500, error message rendered as an HTML
fragment suitable for HTMX `hx-target` swap.

See [data-formats.md](data-formats.md) for the full CSV specification.

---

### `DELETE /data/{symbol}`

Delete all price records for the given symbol. Triggered by HTMX
`hx-delete` on the data manager table row.

**Response**: HTTP 200, empty body. HTMX removes the table row from the DOM.

---

### `GET /experiments`

Renders a table of all experiments with their names, models, horizon, and most
recent run status.

---

### `GET /experiments/new`

Renders the experiment creation form, pre-populated with default values.
Requires at least one symbol to have been uploaded.

---

### `POST /experiments`

Create a new experiment. Optionally kick off an immediate simulation run.

**Request**: `application/x-www-form-urlencoded`

| Field                  | Type     | Required | Default  | Description                                           |
|------------------------|----------|----------|----------|-------------------------------------------------------|
| `name`                 | string   | yes      | —        | Human-readable experiment name                        |
| `description`          | string   | no       | `""`     | Optional description                                  |
| `symbols`              | string[] | yes      | —        | Repeated field; one value per asset (e.g. `AAPL`)     |
| `weights`              | float[]  | no       | equal    | Repeated field; one value per symbol. If omitted or unparseable, equal weights are used. |
| `num_paths`            | int      | yes      | —        | Number of Monte Carlo paths                           |
| `horizon_days`         | int      | yes      | —        | Simulation horizon in trading days                    |
| `lookback_days`        | int      | yes      | —        | Historical lookback window in trading days            |
| `start_value`          | float    | yes      | —        | Starting portfolio value (dollars)                    |
| `model`                | string   | yes      | —        | `"gbm"` or `"bootstrap"`                             |
| `annual_contribution`  | float    | no       | `0`      | Annual cash contribution (dollars)                    |
| `run_now`              | string   | no       | `""`     | Set to `"1"` to immediately queue a simulation run    |

**Success response**: redirect to `/experiments/{id}` (or `/runs/{run_id}` if
`run_now=1`).

---

### `GET /experiments/{id}`

Renders the experiment detail page: configuration summary, list of past runs,
and a button to trigger a new run.

| URL parameter | Description            |
|---------------|------------------------|
| `{id}`        | Experiment ID (int64)  |

---

### `POST /experiments/{id}/run`

Trigger a new simulation run for an existing experiment.

**Request**: no body required (form submit or HTMX `hx-post`).

**Response**: redirect to `/runs/{run_id}` for the new run.

---

### `GET /runs/{id}`

Renders the simulation results page for a completed run, including:

- Percentile fan chart (P5 / P25 / P50 / P75 / P95) rendered with Chart.js
- Summary statistics table (mean, standard deviation, probability of loss,
  median max drawdown, P95 max drawdown, median CAGR)
- Run metadata (model, paths, horizon, seed)

| URL parameter | Description       |
|---------------|-------------------|
| `{id}`        | Run ID (int64)    |

---

### `GET /static/*`

Static assets served directly from the `web/static/` directory.

| Path                          | Description                        |
|-------------------------------|------------------------------------|
| `/static/css/main.css`        | Application styles                 |
| `/static/js/charts.js`        | Chart.js integration               |
| `/static/vendor/htmx.min.js`  | HTMX 2.0.4                        |
| `/static/vendor/chart.min.js` | Chart.js 4.4.2                    |

---

## HTMX Integration

The frontend uses HTMX for partial-page updates. Key patterns:

| HTMX attribute / header       | Where used                              | Behaviour                                         |
|-------------------------------|-----------------------------------------|---------------------------------------------------|
| `hx-post="/data/upload"`      | CSV upload form                         | Submits multipart form; response triggers redirect |
| `HX-Redirect: /data`          | Server → client on upload success       | HTMX performs client-side redirect                |
| `hx-delete="/data/{symbol}"`  | Delete button on data manager table     | Removes the table row on 200 response             |
| `hx-post="/experiments/{id}/run"` | Run button on experiment detail     | Triggers simulation; follows redirect to results  |

---

## Error Handling

All error responses render an HTML fragment with an human-readable message.
HTTP status codes follow standard conventions:

| Status | Meaning                                          |
|--------|--------------------------------------------------|
| 400    | Bad request — missing or invalid form field      |
| 404    | Experiment or run not found                      |
| 500    | Internal server error (logged via `slog`)        |
