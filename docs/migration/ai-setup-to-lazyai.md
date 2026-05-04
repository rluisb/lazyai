# Migration from ai-setup to LazyAI

LazyAI is the new Go-only identity for the former `ai-setup` project. This restructure is intentionally breaking: the old npm/npx distribution and old binary names are not compatibility aliases in this release.

## Repository and docs

| Old | New |
|---|---|
| `github.com/ricardoborges-teachable/ai-setup` | `github.com/rluisb/lazyai` |
| `https://ricardoborges-teachable.github.io/ai-setup/` | `https://rluisb.github.io/lazyai/` |
| `@ricardoborges-teachable/ai-setup` | No npm package; use Go installs |

## Install commands

Replace npm/npx usage with Go installs:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

## Submodule version tags

LazyAI packages are independent Go submodules. Release tags must be prefixed with the module directory:

| Package | Tag format |
|---|---|
| CLI | `packages/cli/vX.Y.Z` |
| Orchestrator | `packages/orchestrator/vX.Y.Z` |
| Diffviewer | `packages/diffviewer/vX.Y.Z` |

Root `vX.Y.Z` tags do not version these submodules for `go install`.

## Command renames

| Old command | New command |
|---|---|
| `ai-setup` | `lazyai-cli` |
| `ai-setup-orchestrator` | `lazyai-orchestrator` |
| `diffviewer` | `lazyai-diffviewer` |

Examples:

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

## MCP config changes

Generated MCP configs should reference `lazyai-orchestrator`, not `ai-setup-orchestrator` or npm package names.

If you have hand-edited MCP configuration, update the command manually. For example:

```json
{
  "command": "lazyai-orchestrator",
  "args": ["connect"]
}
```

Then re-run:

```bash
lazyai-cli compile
lazyai-cli doctor
```

## Local state and data directories

Existing managed projects may still contain files such as `.ai-setup.json`, `.ai-setup.db`, `.ai-setup.toml`, or `.ai-setup-backup/`. Those local file names are not automatically renamed by this package restructure.

Existing orchestrator runtime data under old user data directories is also not migrated automatically. If you need historical runtime state, copy or archive it manually before switching generated MCP entries to `lazyai-orchestrator`.

## What was removed

- npm/npx install and execution paths
- Node/pnpm workspace development commands
- Old binary names as shipped commands
- Old repository owner/package identity in official docs and Pages settings

## Recommended migration checklist

1. Install `lazyai-cli` with `go install`.
2. Install `lazyai-orchestrator` if you use the orchestrator MCP server.
3. Replace shell scripts and docs references from `ai-setup` to `lazyai-cli`.
4. Replace MCP command references from `ai-setup-orchestrator` to `lazyai-orchestrator`.
5. Run `lazyai-cli update --check`, then `lazyai-cli update` when ready.
6. Run `lazyai-cli doctor` to verify generated files and drift.
