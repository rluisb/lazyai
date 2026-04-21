# Task 013 — MCP CLI integration test (fake runner)

**Phase:** 4 (verification)
**Estimated LOC:** ~100

## Goal

Using the injectable `ClaudeCLIRunner` from Task 008, write tests that pin the shape of `claude mcp add-json` calls per scope and that exercise both happy paths and fallback paths.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/claudecode_cli_test.go` (new) | Define a `recordingRunner` that captures every `Run(ctx, cwd, args...)` call. Cases: (a) project scope → assert `add-json <name> <json> -s project` with `cwd = <project>`; (b) global scope → assert `-s user`; (c) workspace scope → assert `-s project` with `cwd = <workspace>`; (d) pre-existing server → `mcp get` succeeds, `add-json` NOT called; (e) CLI missing → direct-write path runs, warning emitted. |

## Acceptance criteria

- [ ] Every scope-to-flag mapping is pinned in a test (project → `-s project`, global → `-s user`, workspace → `-s project`)
- [ ] JSON payload shape is pinned via a golden fixture (stdio, http, and sse server types)
- [ ] Fallback path is exercised and verified (runner returns `LookupClaudeBinary` = false OR runner returns error)
- [ ] Warning line is asserted (captured via a test logger or `io.Writer`)

## Test plan

- Self-verifying.
- Fixture `testdata/mcp_canonical.jsonc` with one stdio, one http, one sse server.
- Golden files per scope: `testdata/expected_add_json_project.txt`, `_user.txt`, `_workspace.txt`.

## Dependencies

- Tasks 008, 009, 010.
