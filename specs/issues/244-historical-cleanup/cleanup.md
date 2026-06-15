# Issue 244 Historical Cleanup

## Final classification

| Item | Final status | Action |
|---|---|---|
| `packages/cli/library/agents/*` | Removable archive/historical material | Moved to `archive/issue-244-historical-library/packages/cli/library/agents/`. Active adapters use `packages/cli/library/canonical/agents/`. |
| `packages/cli/library/copilot/agents/*` | Removable archive/historical material | Moved to `archive/issue-244-historical-library/packages/cli/library/copilot/agents/`. Copilot output is generated from canonical markdown. |
| `packages/cli/library/skills/orchestrate.md` | Removable archive/historical material | Moved to `archive/issue-244-historical-library/packages/cli/library/skills/orchestrate.md`. Active skills are canonical or setup-core markdown, not `@ai-setup/orchestrator` MCP guidance. |
| `packages/cli/library/standards/starter/orchestration-patterns.md` | Removable archive/historical material | Moved to `archive/issue-244-historical-library/packages/cli/library/standards/starter/orchestration-patterns.md`. Starter standards no longer seed orchestration runtime guidance. |
| `packages/cli/cmd/update.go` legacy agent cleanup | Compatibility-only | Kept. It deletes only unmodified library-owned pre-canonical agent files and preserves user-owned or user-edited files; covered by `packages/cli/cmd/update_test.go`. |
| User-provided legacy MCP entries named `orchestrator` | Compatibility-only | Kept in compiler preservation tests. The default MCP catalog excludes `orchestrator`; user-owned `.ai/mcp.json` entries still round-trip. |
| `packages/cli/internal/runtime/schema.go` historical `SchemaV1` | Archived out of production code | Removed from production runtime schema. Runtime migration tests now use a minimal local legacy fixture. |
| `packages/cli/internal/db/migrations.go` migrations 009-011 | Migration/rollback-only | Kept unchanged because package rules prohibit migration edits without explicit human approval. They remain setup-store history, not default adapter output. |

## Default-output checks

- `packages/cli/internal/adapter/adapter_contract_test.go` asserts OpenCode, Claude Code, and Copilot install `primary-agent` and do not emit `orchestrator` agent files or `orchestrate` skill/agent output.
- `packages/cli/internal/adapter/mcp_compiler_test.go` asserts `packages/cli/library/mcp/catalog.json` does not ship a default `orchestrator` MCP server.
- `packages/cli/cmd/command_excision_test.go` continues to assert retired `task`, `workflow`, `orchestration`, `mcp-setup`, and `eval` root command surfaces are absent.

## Migration constraint

`packages/cli/internal/db/migrations.go` was intentionally not edited. Safe removal of old setup-store migrations requires an explicit migration-retention decision because existing `.ai-setup` stores may rely on the version history.