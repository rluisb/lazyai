# Task 004 — MCP per-server deep-merge on compile

**Phase:** 2
**Status:** ✅ complete (2026-04-19)
**Depends on:** 001

## Implementation Notes

- `compileOpenCodeMCP` (`internal/adapter/mcp_compiler.go`) no longer overwrites the whole `mcp` key. It now delegates to a new `mergeOpenCodeMcpServers(existingRaw, managed)` helper that:
  - Keeps any server name from the pre-existing config that is NOT in the managed catalog (user-authored servers).
  - Upserts every entry from the managed catalog (so toggling enabled/disabled propagates).
  - On name collision between user-authored and managed, the managed entry wins — documented inline.
- Two new tests in `mcp_compiler_test.go`:
  - `TestCompileOpenCodeMCP_PreservesUserAuthoredServer` — pre-seeds `opencode.jsonc` with a `userAuthored` server and asserts it survives compile alongside the managed `memory` server.
  - `TestCompileOpenCodeMCP_ManagedWinsOnNameCollision` — pre-seeds a user `memory` with a custom command and verifies the managed definition overwrites it.

## Verification

- `go test ./... -count=1` — PASS
- `go vet ./...` — clean

## Scope

Change `compileOpenCodeMCP` from whole-`mcp`-key overwrite to per-server upsert. User-authored MCP servers (not present in the current ai-setup catalog snapshot) survive compile cycles.

## Changes

- `internal/adapter/mcp_compiler.go#compileOpenCodeMCP`:
  - Read existing `mcp` map from `opencode.jsonc`.
  - Build `managedKeys := set(names of servers in current catalog)`.
  - For each key in `ocMcp` (managed servers): overwrite existing[key] = ocMcp[key].
  - For each key in existing: if `key ∉ managedKeys` → preserve.
  - Write merged map back.
- Document the "managed by current catalog snapshot" rule inline as a short comment.

## Tests

- `internal/adapter/mcp_compiler_test.go`:
  - Pre-seed config with `mcp.userServer = {type:"local",command:["foo"]}` and compile → `userServer` still present; managed servers present.
  - Toggle a managed server disabled → recompile → managed server updated, `userServer` untouched.
  - Empty starting config → compile produces same output as current behavior (regression guard).
  - Conflict case: user-authored server with same name as managed → managed wins (documented limit).

## Definition of Done

- AC-4 checklist items all green.
- All existing MCP compile tests still pass.
