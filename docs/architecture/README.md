# architecture/

How the system is built **today** — the current shape of the code, intended as a cold-start reference for an agent or engineer dropping into the repo.

**Put here:**
- System-overview docs that describe layers, packages, and dependency flow as they are right now.
- Diagrams and prose that explain the present architecture and would still be true a week later.

**Do not put here:**
- Proposals for future architecture — `design/`.
- Phased migration sequencing — `plans/`.
- Vendor or integration API quirks — `reference/`.
- Runbooks — `operations/`.

**Naming convention:** `<yyyy-mm-dd>-<topic>.md`
Examples: `2026-05-02-overview.md`, `2026-08-15-storage-layer.md`.

**Allowed `status:` values:** `Stable`, `Superseded`.

When the architecture changes materially, supersede the existing doc with a new one and update `superseded_by:` on the old one.
