# Contributing to Drift

Thank you for contributing. This document covers code design principles, quality
standards, and the required branch / pull-request workflow.

---

## Table of Contents

1. [Architecture & design principles](#architecture--design-principles)
2. [Code quality standards](#code-quality-standards)
3. [Testing requirements](#testing-requirements)
4. [Branch & PR workflow](#branch--pr-workflow)
5. [Commit conventions](#commit-conventions)
6. [Development setup](#development-setup)

---

## Architecture & design principles

Drift uses **hexagonal architecture** (ports and adapters). Keep this boundary
sharp — domain logic must never import adapters, and adapters must never contain
business logic.

```
cmd/drift/          ← composition root; wires everything together
internal/
  domain/           ← pure business types and rules; zero external deps
  ports/
    inbound/        ← service interfaces called by HTTP handlers
    outbound/       ← repository interfaces called by services
  services/         ← application logic; depends only on ports + domain
  adapters/
    http/           ← chi router, handlers, templates, static files
    ingestion/      ← CSV / JSON parsing
    storage/sqlite/ ← SQLite implementation of outbound ports
```

### Rules

| Rule | Detail |
|---|---|
| **Domain is king** | All domain types live in `internal/domain`. No framework imports. |
| **Depend inward** | Adapters depend on ports; ports depend on domain; domain depends on nothing. |
| **No leaky abstractions** | Services receive and return domain types only — never DB rows or HTTP payloads. |
| **Immutable domain values** | Prefer passing structs by value for domain types; use pointers only for optional / nullable fields. |
| **Error wrapping** | Always wrap errors with context: `fmt.Errorf("get experiment: %w", err)`. |
| **No global state** | Use dependency injection via constructor functions (`NewXxxService`). |
| **Context plumbing** | Every function that touches I/O must accept and propagate `context.Context` as its first argument. |

---

## Code quality standards

All code merged to `main` must satisfy **all** of the following.

### Formatting

```bash
make fmt   # runs go fmt ./...
```

The CI `format` job will fail on any unformatted file. Do not disable this check.

### Vetting

```bash
make vet   # runs go vet ./...
```

Fix all `go vet` findings before opening a PR.

### Linting

```bash
make lint  # runs golangci-lint run ./...
```

Active linters: `govet`, `errcheck`, `staticcheck`, `revive`, `ineffassign`,
`unused`, `misspell`, `gocritic`. Configuration lives in `.golangci.yml`.

- **Do not use bare `//nolint`** — always name the linter and add a `//` comment
  explaining why the suppression is justified:
  ```go
  defer rows.Close() //nolint:errcheck // Close in defer; error captured by rows.Err()
  ```
- Fix the root cause instead of suppressing whenever possible.

### Doc comments

Every **exported** symbol must have a Go doc comment:

```go
// ComputeStats derives ResultStats from the completed set of simulated paths.
func ComputeStats(paths []SimulatedPath, startValue, horizonYears float64) ResultStats { … }
```

Comments on `const` blocks must appear on the block, not each constant.

### Package names

Use `lowercase`, never `mixedCaps`. Package names must match the last element of
the import path (e.g. `package httpadapter` for `internal/adapters/http`).

---

## Testing requirements

| Requirement | Detail |
|---|---|
| **Race detector** | All tests run with `-race`. Never skip this. |
| **Table-driven** | Prefer table-driven tests with `t.Run(name, …)` for coverage of edge cases. |
| **No global test state** | Each test must be independent; use `t.TempDir()` or `:memory:` for SQLite. |
| **Arrange-Act-Assert** | Structure test bodies clearly; add blank lines between sections. |
| **Assert early** | Use `t.Fatalf` (not `t.Errorf`) when a test cannot continue after a failure. |
| **No sleeps** | Never use `time.Sleep` in tests; use channels or context cancellation. |
| **Coverage target** | Domain and service packages should stay above 80 % coverage. |

```bash
make test                         # run all tests with race detector
go test -race -run TestFoo ./...  # run a single test
go test -race -cover ./...        # print coverage summary
```

White-box tests (package `services`, private helpers) live in the same package.
Black-box integration tests live in `_test` packages or `internal/adapters/*`.

---

## Branch & PR workflow

This project enforces a **topic-branch → PR → squash-merge** workflow. Direct
pushes to `main` are not allowed.

### Branch naming

```
<type>/<short-description>
```

| Type | When to use |
|---|---|
| `feat/` | New feature or capability |
| `fix/` | Bug fix |
| `refactor/` | Code restructure with no behaviour change |
| `test/` | Adding or improving tests |
| `docs/` | Documentation only |
| `ci/` | CI / tooling changes |
| `chore/` | Dependency bumps, minor housekeeping |

Examples: `feat/bootstrap-model`, `fix/csv-adjclose-empty`, `ci/release-workflow`

### Opening a PR

1. **Branch off `main`**: `git checkout -b feat/my-feature main`
2. **Keep it small**: one logical change per PR. Large refactors should be split.
3. **Fill in the PR template** completely (checklist, description, screenshots if UI changes).
4. **CI must be green** before requesting review — don't open a PR with a known failing check.
5. **Self-review first**: read your own diff in the GitHub UI before assigning a reviewer.
6. **Request review** from at least one other contributor (see `CODEOWNERS`).

### Review etiquette

- Reviewers must respond within **2 business days**.
- Use **Conventional Comments** prefixes: `nit:`, `question:`, `suggestion:`, `blocker:`.
- A `blocker:` comment must be resolved before merging.
- Authors **must not dismiss** a review — resolve the comment or reply explaining why you disagree.

### Merging

- **Squash-merge only** — keeps `main` history linear and readable.
- The squashed commit message must follow the [commit conventions](#commit-conventions) below.
- Delete the branch after merging.

### Hotfixes

For urgent production fixes:

1. Branch from the latest release tag: `git checkout -b fix/critical-bug v1.2.3`
2. Fix, test, open PR to `main`.
3. Tag a patch release after merge.

---

## Commit conventions

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>

[optional body]

[optional footers — e.g. Fixes #42]
```

| Type | When |
|---|---|
| `feat` | New user-visible feature |
| `fix` | Bug fix |
| `refactor` | Code change with no behaviour change |
| `test` | Tests only |
| `docs` | Documentation only |
| `ci` | CI / build system |
| `chore` | Dependency bump, minor housekeeping |
| `perf` | Performance improvement |

Scope is the affected package or area, e.g. `(domain)`, `(sqlite)`, `(http)`.

**Bad**: `fixed stuff`, `WIP`, `more changes`
**Good**: `fix(csv): skip rows with missing adjusted_close`

---

## Development setup

```bash
# Prerequisites: Go >= 1.25, golangci-lint >= 2.x, gh CLI

git clone git@github.com:gjcourt/drift.git
cd drift

go mod download

make build   # compile
make test    # run tests
make lint    # lint
make check   # fmt + vet + lint + test (full pre-push check)
```

Override defaults with environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `DRIFT_ADDR` | `:8080` | HTTP listen address |
| `DRIFT_DB` | `drift.db` | SQLite database path |
| `DRIFT_TMPL_DIR` | auto-detected from source | Template directory |
| `DRIFT_STATIC_DIR` | auto-detected from source | Static assets directory |
