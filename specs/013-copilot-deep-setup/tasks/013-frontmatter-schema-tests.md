# Task 013 — Frontmatter + schema tests for emitted Copilot artifacts

**Phase:** 6 (tests)
**Estimated LOC:** ~120

## Goal

Parse every emitted Copilot artifact after install and assert required fields. Catches regressions in the transform code (task 005 skills→agents) and in library content (tasks 001, 002).

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/copilot_frontmatter_test.go` (new) | Run Install at project scope, walk `.github/agents/*.agent.yaml`, unmarshal each, assert: `name` non-empty lowercase; `prompt` non-empty; `description` non-empty. |
| Same file | Walk `.github/instructions/*.instructions.md`, parse frontmatter via `internal/frontmatter`, assert `applyTo` present and non-empty. |
| Same file | Walk `.github/chatmodes/*.chatmode.md`, assert `description` present. |
| Same file | Run compile at global scope, parse `~/.copilot/mcp-config.json`, assert `mcpServers` top-level key, each entry has either `command` (stdio) or `url` (http/sse). |

## Acceptance criteria

- [ ] Parse-every-agent test asserts schema at both project and global scope
- [ ] Instructions test asserts `applyTo` globs are non-empty strings
- [ ] MCP user-scope shape test asserts canonical `"type": "stdio"` (not `"local"`)
- [ ] Tests use the real library (embed FS) — not stubs — so library regressions surface here

## Test plan

Self-contained in the new test file. Reuse `t.TempDir()` for HOME + project dir.

## Notes

- If adding a YAML dependency, coordinate with task 003 to use the same library (no split).
- Keep the assertions minimal — this is contract-level, not behavioral.
