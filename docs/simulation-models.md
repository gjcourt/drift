# Simulation Models

Drift supports two families of stochastic models for generating Monte Carlo price paths.

---

## Shared concepts

### Trading days

All time horizons are expressed in **trading days** (approximately 252 per calendar year). The simulation produces `HorizonDays + 1` values per path: index 0 is `StartValue`, index `HorizonDays` is the terminal value.

### Portfolio compounding

For multi-asset portfolios, each day's portfolio value is the **weighted sum** of the individual asset values, using the weights from `Portfolio.Assets`. The current implementation assumes static weights (no intra-period rebalancing).

### Annual contributions

If `SimulationConfig.AnnualContribution != 0`, the amount is added to the path value at every 252nd day.

### Reproducibility

Set `SimulationConfig.Seed` to a non-nil `*int64` to get deterministic output. Two runs with the same seed, model, and config will produce identical paths. Omit the seed (leave `nil`) for non-deterministic behaviour.

---

## Geometric Brownian Motion (GBM)

**Model identifier**: `"gbm"`

### What it models

GBM is the canonical continuous-time model for equity prices. Log-returns are assumed to be normally distributed and independently and identically distributed (i.i.d.) over time. This corresponds to the Black-Scholes assumption.

### Parameter estimation

For each asset, GBM parameters are estimated from the historical `LookbackDays` of adjusted-close prices:

$$\mu_{\text{annual}} = \bar{r} \cdot 252$$

$$\sigma_{\text{annual}} = \sigma_r \cdot \sqrt{252}$$

where $\bar{r}$ is the mean of the daily log-returns $r_t = \ln(P_t / P_{t-1})$ and $\sigma_r$ is their standard deviation.

### Path generation

Each daily step uses the exact GBM discretisation:

$$S_{t+1} = S_t \cdot \exp\!\left[\left(\mu - \tfrac{1}{2}\sigma^2\right)\Delta t + \sigma\sqrt{\Delta t}\, Z_t\right]$$

where $\Delta t = 1/252$ and $Z_t \sim \mathcal{N}(0,1)$.

Applied to the portfolio:

$$V_{t+1} = V_0 \cdot \sum_{i} w_i \cdot \prod_{s=1}^{t+1} \exp\!\left[\left(\mu_i - \tfrac{1}{2}\sigma_i^2\right)\Delta t + \sigma_i\sqrt{\Delta t}\, Z_{i,s}\right]$$

### Strengths

- Analytically tractable; fast to simulate.
- Well-understood statistical properties.

### Limitations

- Assumes log-normal returns; does not capture fat tails, volatility clustering, or mean reversion.
- Parameters are estimated from a fixed lookback window and may not reflect future regimes.

---

## Historical Bootstrap

**Model identifiers**: `"bootstrap"` and `"block_bootstrap"`

### What it models

Bootstrap simulation re-samples from the **empirical distribution** of historical daily log-returns rather than assuming a parametric distribution. This naturally captures fat tails, skewness, and any other non-normality present in the data.

### Data used

For each asset, the adapter fetches up to `LookbackDays + 1` price records, computes daily log-returns, and stores them as a slice. Returns are filtered to remove any pair where either price is zero or non-positive.

### Path generation

Each daily step draws one log-return per asset **uniformly at random with replacement** from the stored historical returns:

$$V_{t+1} = V_t \cdot \exp\!\left(\sum_{i} w_i \cdot r_{i,t}\right)$$

where $r_{i,t}$ is a randomly selected historical log-return for asset $i$.

> **Current implementation note**: `"block_bootstrap"` is accepted as a model
> identifier but currently uses the same i.i.d. sampling as `"bootstrap"`. True
> block bootstrap (preserving autocorrelation by drawing contiguous windows of
> returns) is planned for a future release.

### Strengths

- No distributional assumption; empirical fat tails and asymmetry included automatically.
- Straightforward to implement and explain.

### Limitations

- Assumes returns are exchangeable (i.i.d.); does not capture serial correlation or volatility clustering.
- Quality degrades with small lookback windows (fewer unique return observations).

---

## Choosing a model

| Consideration | GBM | Bootstrap |
|---|---|---|
| Short simulation horizon (< 1 year) | Good | Good |
| Long horizon (> 10 years) | Reasonable | Good |
| Fat-tail sensitivity | Poor | Good |
| Requires large history | No | Yes (≥ 252 days recommended) |
| Interpretable parameters | Yes (μ, σ) | No |
| Speed | Very fast | Fast |

---

## Result statistics

After all paths are generated, `domain.ComputeStats` computes:

| Statistic | Description |
|---|---|
| `P5`, `P25`, `P50`, `P75`, `P95` | Portfolio value at the given percentile at horizon |
| `Mean` | Arithmetic mean of terminal values |
| `StdDev` | Standard deviation of terminal values |
| `ProbabilityOfLoss` | Fraction of paths where terminal value < start value |
| `MedianMaxDrawdown` | Median of per-path maximum drawdown (negative fraction) |
| `P95MaxDrawdown` | 95th-percentile worst drawdown |
| `MedianCAGR` | Median compound annual growth rate: $(V_T / V_0)^{1 / T} - 1$ |
