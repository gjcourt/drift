# Data Formats

This document describes the file formats that Drift accepts for price data and
experiment configuration.

---

## CSV Price Data

Drift accepts CSV files containing historical price data. Two modes are
supported: **single-symbol** and **multi-symbol**.

### Single-Symbol CSV

When a file contains data for exactly one asset, name the file `SYMBOL.csv`
(e.g. `AAPL.csv`). The symbol is derived from the filename (uppercased, without
the extension). No `symbol` column is required.

**Required columns** (header names are case-insensitive, leading/trailing
whitespace is trimmed):

| Column           | Type    | Notes                                                  |
|------------------|---------|--------------------------------------------------------|
| `date`           | string  | ISO 8601 format: `YYYY-MM-DD`                          |
| `adjusted_close` | float   | Split- and dividend-adjusted closing price. **Required**; rows where this value is missing, empty, or ≤0 are silently dropped. |

**Optional columns** (parsed when present, ignored when absent):

| Column  | Type    | Notes                       |
|---------|---------|-----------------------------|
| `open`  | float   | Opening price               |
| `high`  | float   | Daily high                  |
| `low`   | float   | Daily low                   |
| `close` | float   | Unadjusted closing price    |
| `volume`| integer | Trading volume               |

**Example** (`AAPL.csv`):

```csv
date,open,high,low,close,adjusted_close,volume
2023-01-03,130.28,130.90,124.17,125.07,124.50,112117500
2023-01-04,126.89,128.66,125.08,126.36,125.78,89113600
2023-01-05,127.13,127.77,124.76,125.02,124.44,80962700
```

### Multi-Symbol CSV

When a file contains data for multiple assets, include a `symbol` column. The
filename can be anything (e.g. `portfolio.csv`).

**Required columns**: `date`, `adjusted_close`, **`symbol`**

All optional columns from the single-symbol format are also supported.

**Example** (`portfolio.csv`):

```csv
date,symbol,adjusted_close,volume
2023-01-03,AAPL,124.50,112117500
2023-01-03,MSFT,238.51,27058100
2023-01-04,AAPL,125.78,89113600
2023-01-04,MSFT,231.93,30074500
```

### Skip Behaviour

Rows are silently skipped when:
- `adjusted_close` is empty, non-numeric, or ≤0
- `date` is missing or does not parse as `YYYY-MM-DD`

All other rows are imported. Duplicate `(symbol, date)` pairs are accepted;
the last row seen wins at the storage layer.

---

## JSON Experiment Configuration

Experiments can be staged programmatically by `POST`-ing a JSON document to
`/experiments` with `Content-Type: application/json`, or by using the web UI
form. The JSON schema mirrors the `ExperimentConfig` struct in
`internal/adapters/ingestion/json.go`.

### Schema

```jsonc
{
  "version": "1",                        // string, required. Use "1".

  "experiment": {
    "name":        "My Retirement Plan", // string, required
    "description": "30-year projection"  // string, optional
  },

  "portfolio": {
    "assets": [
      { "symbol": "AAPL", "weight": 0.6 }, // symbol: string (uppercase), weight: float
      { "symbol": "MSFT", "weight": 0.4 }
    ],
    "rebalance": "monthly"  // "none" | "daily" | "monthly" | "yearly" (default: "none")
  },

  "simulation": {
    "model":         "gbm",   // "gbm" | "bootstrap" (default: "gbm")
    "num_paths":     1000,    // int, number of Monte Carlo paths
    "horizon_days":  7560,    // int, simulation horizon in trading days (252/yr)
    "lookback_days": 1260,    // int, historical window for parameter estimation
    "start_value":   100000,  // float, starting portfolio value in dollars
    "seed":          42       // int64 | null — null means non-deterministic
  },

  "parameters": {
    "annual_contribution": 12000, // float, dollars added each year (default: 0)
    "withdrawal_rate":     0.04   // float | null, fraction withdrawn per year (default: null → 0)
  }
}
```

### Field Reference

#### `portfolio.rebalance`

| Value     | Meaning                                         |
|-----------|-------------------------------------------------|
| `"none"`  | No rebalancing; weights drift with market moves |
| `"daily"` | Rebalance to target weights every trading day   |
| `"monthly"` | Rebalance on the first day of each month      |
| `"yearly"` | Rebalance on the first day of each year        |

#### `simulation.model`

| Value         | Description                                                   |
|---------------|---------------------------------------------------------------|
| `"gbm"`       | Geometric Brownian Motion — parametric, assumes log-normality |
| `"bootstrap"` | Empirical bootstrap — samples historical return sequences with replacement |

See [simulation-models.md](simulation-models.md) for full model details.

#### `simulation.seed`

Set to a non-null integer for reproducible results. Omit or set to `null` for
a random seed (useful in production to get varied outcomes each run).

### Example

```json
{
  "version": "1",
  "experiment": {
    "name": "60/40 Portfolio — 20yr Horizon",
    "description": "Standard balanced allocation with annual rebalancing"
  },
  "portfolio": {
    "assets": [
      { "symbol": "SPY",  "weight": 0.60 },
      { "symbol": "AGG",  "weight": 0.40 }
    ],
    "rebalance": "yearly"
  },
  "simulation": {
    "model":         "gbm",
    "num_paths":     5000,
    "horizon_days":  5040,
    "lookback_days": 2520,
    "start_value":   250000,
    "seed":          null
  },
  "parameters": {
    "annual_contribution": 0,
    "withdrawal_rate": null
  }
}
```
