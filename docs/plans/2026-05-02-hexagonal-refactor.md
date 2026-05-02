---
title: "Hexagonal architecture migration"
status: "In progress"
created: "2026-05-02"
updated: "2026-05-02"
updated_by: "george"
tags: ["architecture", "hex", "refactor"]
---

# Hexagonal architecture migration

## Current layout

```
internal/
  domain/            — SimulationConfig, Portfolio, Asset, SimulatedPath, etc.
  ports/
    inbound/         — IngestionPort, SimulationPort, ResultsPort
    outbound/        — AssetRepository, ExperimentRepository, SimulationRepository
  services/          — IngestionService, SimulationService, ResultsService (target: app/)
  adapters/
    http/            — HTTP handlers + templates (driving adapter)
      handlers/
      templates/
    ingestion/       — CSV/file ingestion (driven adapter)
    storage/
      sqlite/        — SQLite-backed repositories
```

The ports split is already done. The main gap is: `services/` should be
`app/`, and `services/ingestion.go` imports `adapters/ingestion` directly
instead of using the `outbound.AssetRepository` port.

## Migration steps

1. **Fix `services/ingestion.go`** — replace the direct import of
   `adapters/ingestion` with the `outbound.AssetRepository` port interface.
   Inject the adapter at the composition root. Removes the blocked depguard
   rule. One PR.

2. **Activate `services-no-adapters` depguard rule** — add after step 1 passes
   CI. One PR (config-only).

3. **Rename `services/` → `app/`** — `git mv internal/services internal/app`.
   Update all import paths. Update depguard rules (replace `services` with
   `app`). One PR.

4. **Add function-field fakes** — add `FakeAssetRepository`,
   `FakeExperimentRepository`, `FakeSimulationRepository` to
   `internal/testdoubles/`, wire into `ServerDeps`.

5. **Refactor app-layer tests** to use `testdoubles.NewServerDeps()` instead
   of ad-hoc in-test stubs.

## Depguard notes

Bootstrap rules active: `domain-no-other-internal`, `ports-no-impl`,
`adapters-isolation`.

Pending rule (blocked): `services-no-adapters` — `services/ingestion.go`
imports `adapters/ingestion` directly. Unblocked after step 1.
