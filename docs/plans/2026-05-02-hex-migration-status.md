---
title: "Hex migration status"
status: "In progress"
created: "2026-05-02"
updated: "2026-05-02"
updated_by: "george"
tags: ["architecture", "hex", "tracking"]
---

# Hex migration status

## Depguard rules

| Rule | Status | Notes |
|---|---|---|
| `domain-no-other-internal` | Active ✓ | Domain has no internal imports |
| `ports-no-impl` | Active ✓ | Ports only import domain |
| `adapters-isolation` | Active ✓ | Adapters don't cross-import or call app |
| `app-no-adapters` | Active ✓ | `app/ingestion.go` refactored to use `outbound.CSVParser` port |

## Migration checklist

- [x] Step 1 — refactor `services/ingestion.go` to use `outbound.CSVParser` port (wired via `ingestion.Parser{}` in `main.go`)
- [x] Step 2 — activate `app-no-adapters` depguard rule
- [x] Step 3 — rename `services/` → `app/`
- [ ] Step 4 — add fakes to `testdoubles/`, wire `ServerDeps`
- [ ] Step 5 — migrate app-layer tests to use `testdoubles.NewServerDeps()`
