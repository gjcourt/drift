# reference/

Information you look things up in — APIs, domain knowledge, integration specs, configuration tables, math models.

**Put here:**
- HTTP route reference and request/response shapes.
- Data format specifications (CSV/JSON layout, field meanings).
- Algorithm / model reference (math definitions, parameter semantics).
- Long-shelf-life lookup material that an engineer or agent re-reads each time.

**Do not put here:**
- Runbook steps — `operations/`.
- Architecture overview — `architecture/`.
- Spike output — `research/`.

**Naming convention:** `<yyyy-mm-dd>-<topic>.md`
Examples: `2026-05-02-api.md`, `2026-05-02-data-formats.md`, `2026-05-02-simulation-models.md`.

**Allowed `status:` values:** `Stable`, `Superseded`.

Date prefix is bumped when the doc is materially revised.
