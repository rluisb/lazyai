# Research: Dashboard Run Read-Model Extraction

## Current State

`ReadModel` in `packages/orchestrator/internal/dashboard/readmodel.go` has 3 remaining direct `database *db.DB` calls:

| Method | Line | Query pattern |
|--------|------|---------------|
| `ListRuns` | 143 | `m.database.QueryContext(ctx, query, args...)` — UNION ALL across chain/team/workflow runs with filtering, search, attention, pagination |
| `runCountsByState` | 255 | `m.database.QueryContext(ctx, ...)` — state aggregation via UNION ALL, GROUP BY state |
| `getRunRow` | 473 | `m.database.QueryRowContext(ctx, query, id)` — single run lookup by kind + id, with table dispatch |

Everything else is already behind ports.

## Design: Single `RunReadStore` Port

Extracting 3 separate ports would over-fragment the read surface since all three read from the same underlying tables (chain_runs, team_runs, workflow_runs). A single cohesive port is appropriate:

**`ports.RunReadStore`:**

- `ListRuns(ctx, filter RunListFilter) (RunListPage, error)`  
  Replaces the UNION ALL query with filtering, search, attention, HasErrors, pagination. Returns items + optional next cursor. Adapter handles the UNION, WHERE clause construction, LIMIT/OFFSET, and row parsing.

- `CountRunsByState(ctx) (map[string]int, error)`  
  Replaces `runCountsByState`. Returns a map of state → count across all three run kinds.

- `FindRunRow(ctx, kind RunKind, id string) (RunRow, error)`  
  Replaces `getRunRow`. Returns a single run row with all columns. Adapter dispatches to the correct table.

**Domain types needed:**

- `RunListFilter` — mirrors current `RunListOptions` shape (kind, state, search, attention, hasErrors, limit, cursor)
- `RunListPage` — items + next cursor
- `RunRow` — mirrors current `runRow` struct (kind, id, definition_name, definition_version, state, current, project_root, state_json, created_at, updated_at)

## Impact Assessment

| Affected file | Change |
|---------------|--------|
| `ports/run_read_store.go` | New port definition |
| `domain/run_read.go` | New domain filter/page/row types |
| `adapters/sqlite/run_read_store.go` | SQLite adapter implementing the port |
| `adapters/sqlite/run_read_store_test.go` | Adapter tests |
| `internal/dashboard/readmodel.go` | Replace `database *db.DB` with `ports.RunReadStore`. Remove `runUnionQuery`, `scanRunRows`, `scanRunRow`, `getRunRow`, `runTable` and related direct SQL. |
| `internal/dashboard/readmodel_test.go` | Replace `database *db.DB` test seeding with adapter-based test doubles. |
| `internal/dashboard/handlers_test.go` | Update `newDashboardHTTPHandler` to wire the new port. |
| `cmd/lazyai-orchestrator/main.go` | Wire `NewReadModel` with `sqliteadapter.NewRunReadStore(database)`. |

## Risks

- **Pagination contract must be preserved.** Cursor-based offset pagination, limit bounds, and hasMore detection must remain identical.
- **Search escaping must be preserved.** `likeEscape` replaces `%`, `_`, and `\`.
- **UNION ALL ordering must be preserved.** `ORDER BY updated_at DESC, id DESC` across the union.
- **Test seeding currently uses direct DB.** Tests that `seedRun` / `seedEvent` / `seedError` must adapt to use the SQLite adapter internally for seeding or provide a testing adapter.

## Recommended Approach

1. Extract into a single `RunReadStore` port — cohesive, not fragmented.
2. Move all SQL (UNION, filtering, search, attention) into the SQLite adapter.
3. `ReadModel` becomes a pure coordinator calling port methods and composing summaries, budgets, events.
4. Remove `database *db.DB` from `ReadModel` entirely.
5. Seed tests through the adapter internally or provide a lightweight in-memory test adapter.
