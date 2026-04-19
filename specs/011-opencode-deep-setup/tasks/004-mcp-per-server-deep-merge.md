# Task 004 — MCP per-server deep-merge on compile

**Phase:** 2
**Status:** pending
**Depends on:** 001

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
