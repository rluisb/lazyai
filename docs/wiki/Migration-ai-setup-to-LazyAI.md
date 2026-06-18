# Migration from ai-setup to LazyAI

LazyAI is the Go-only identity for the former `ai-setup` project. This restructure is intentionally breaking: old npm/npx paths and old binary names are not compatibility aliases in this release.

## Repository and docs

| Old | New |
|---|---|
| `github.com/rluisb/ai-setup` | `github.com/rluisb/lazyai` |
| `https://rluisb.github.io/ai-setup/` | `https://rluisb.github.io/lazyai/` |
| npm/npx package usage | Go installs with `go install` |

## Install LazyAI commands

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

## Command renames

| Old command | New command |
|---|---|
| `ai-setup` | `lazyai-cli` |
| `ai-setup-orchestrator` | `lazyai-orchestrator` |
| `diffviewer` | `lazyai-diffviewer` |

Example:

```bash
# Before
ai-setup init --enable-servers orchestrator
ai-setup compile
ai-setup doctor

# After
lazyai-cli init --enable-servers orchestrator
lazyai-cli compile
lazyai-cli doctor
```

## MCP configuration changes

Generated MCP configs should reference `lazyai-orchestrator`, not `ai-setup-orchestrator` or npm package names.

If you have hand-edited MCP configuration, update the command manually:

```json
{
  "command": "lazyai-orchestrator",
  "args": ["connect"]
}
```

Then regenerate/check managed files:

```bash
lazyai-cli compile
lazyai-cli doctor
```

## Local state

Existing managed projects may still contain `.ai-setup.json`, `.ai-setup.db`, `.ai-setup.toml`, or `.ai-setup-backup/`. Those local file names are not automatically renamed by this package restructure.

Existing orchestrator runtime data under old user data directories is also not migrated automatically. Archive or copy historical runtime state before switching generated MCP entries if you need it.

## Migration checklist

1. Install `lazyai-cli` with Go.
2. Install `lazyai-orchestrator` if you use the orchestrator MCP server.
3. Replace scripts/docs references from `ai-setup` to `lazyai-cli`.
4. Replace MCP command references from `ai-setup-orchestrator` to `lazyai-orchestrator`.
5. Run `lazyai-cli update --check`, then `lazyai-cli update` when ready.
6. Run `lazyai-cli doctor`.

For detailed docs, use <https://rluisb.github.io/lazyai/migration/ai-setup-to-lazyai/>.
