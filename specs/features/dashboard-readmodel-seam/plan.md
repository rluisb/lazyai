# Implementation Plan: Dashboard Run Read-Model Extraction

## Phase 1: Port and Domain Types

1.  **Create `domain/run_read.go`** — define `RunListFilter`, `RunListPage`, `RunRow`.
2.  **Create `ports/run_read_store.go`** — define `RunReadStore` interface with:
    - `ListRuns(ctx, filter RunListFilter) (RunListPage, error)`
    - `CountRunsByState(ctx) (map[string]int, error)`
    - `FindRunRow(ctx, kind RunKind, id string) (RunRow, error)`

## Phase 2: SQLite Adapter

1.  **Create `adapters/sqlite/run_read_store.go`** — adapter wrapping `db.DB`:
    - Move `runUnionQuery()` into the adapter.
    - Move `ListRuns` SQL (UNION ALL, WHERE clause construction, filtering, search escaping, attention, pagination, row scanning) into the adapter.
    - Move `CountRunsByState` SQL (UNION ALL with GROUP BY state) into the adapter.
    - Move `getRunRow` / `runTable` logic into the adapter as `FindRunRow`.
    - Add compile-time `var _ ports.RunReadStore = (*RunReadStore)(nil)`.
2.  **Create `adapters/sqlite/run_read_store_test.go`** — adapter tests:
    - List with filtering, search, pagination, hasMore detection.
    - Count by state.
    - Find single run row.
    - Not-found behavior.

## Phase 3: Wire ReadModel

1.  **Update `internal/dashboard/readmodel.go`**:
    - Replace `database *db.DB` with `runStore ports.RunReadStore`.
    - Delegate `ListRuns`, `runCountsByState`, `getRunRow` calls to the port.
    - Remove `runUnionQuery`, `scanRunRows`, `scanRunRow`, `getRunRow`, `runTable`, `likeEscape` (moved to adapter).
2.  **Update `NewReadModel`** signature — accept `ports.RunReadStore` instead of `*db.DB`.
3.  **Update `cmd/lazyai-orchestrator/main.go`** — wire `NewRunReadStore(database)`.

## Phase 4: Test Wiring

1.  **Update `internal/dashboard/readmodel_test.go`**:
    - `newDashboardTestReadModel` uses the SQLite adapter instead of direct `*db.DB`.
    - `seedRun`/`seedEvent`/`seedError`/`seedExecutionPlan` helpers continue using `*db.DB` for seeding, but `ReadModel` construction uses the adapter.
    - Add test for injectable `RunReadStore`.
2.  **Update `internal/dashboard/handlers_test.go`**:
    - `newDashboardHTTPHandler` wires the SQLite adapter.

## Execution Order

1.  Domain types + port.
2.  SQLite adapter + adapter tests.
3.  ReadModel wiring.
4.  Test updates + verification.

## Verification Requirements

- `GOWORK=off go test ./adapters/sqlite -count=1` — adapter tests pass.
- `GOWORK=off go test ./internal/dashboard -count=1` — read model tests pass.
- `GOWORK=off go test ./... -count=1` — full orchestrator tests pass.
- `GOWORK=off go build ./cmd/lazyai-orchestrator` — build passes.
- Remove generated binary artifact.
