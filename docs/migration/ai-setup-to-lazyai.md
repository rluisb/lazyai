# Migration from ai-setup to LazyAI

LazyAI is the new Go-only identity for the former `ai-setup` project. This restructure is intentionally breaking: the old npm/npx distribution and old binary names are not compatibility aliases in this release.

## Repository and docs

## Install commands

Replace npm/npx usage with Go installs:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

## Submodule version tags

LazyAI active packages are independent Go submodules. Release tags must be prefixed with the module directory:

| Package     | Tag format                    |
| ----------- | ----------------------------- |
| CLI         | `packages/cli/vX.Y.Z`         |
| Diffviewer  | `packages/diffviewer/vX.Y.Z`  |

Root `vX.Y.Z` tags do not version these submodules for `go install`.

## Command renames

| Old command             | New command                |
| ----------------------- | -------------------------- |
| `ai-setup`              | `lazyai-cli`               |
| `ai-setup-orchestrator` | Removed; no active binary  |
| `diffviewer`            | `lazyai-diffviewer`        |

Examples:

```bash
# Before
ai-setup init --enable-servers filesystem,ai-memory
ai-setup compile
ai-setup doctor

# After
lazyai-cli init --enable-servers filesystem,ai-memory
lazyai-cli compile
lazyai-cli doctor
```

## MCP config changes

Generated MCP configs no longer include the retired orchestration runtime. If you have hand-edited MCP configuration for that server, remove the entry and re-run:

```bash
lazyai-cli compile
lazyai-cli doctor
```

See [Migration: Fortnite / orchestrator removal](fortnite-orchestrator-removal.md) if existing projects still reference the old runtime paths.

## Local state and data directories

Existing managed projects may still contain files such as `.ai-setup.json`, `.ai-setup.db`, `.ai-setup.toml`, or `.ai-setup-backup/`. Those local file names are not automatically renamed by this package restructure.

Historical runtime data under old user data directories is not migrated automatically. Archive or copy it manually if you need it for audit purposes.

## What was removed

- npm/npx install and execution paths
- Node/pnpm workspace development commands
- Old binary names as shipped commands
- Dedicated workflow/task runtime binary and generated MCP entry
- Old repository owner/package identity in official docs and Pages settings

## Recommended migration checklist

1. Install `lazyai-cli` with `go install`.
2. Replace shell scripts and docs references from `ai-setup` to `lazyai-cli`.
3. Remove retired runtime MCP entries from hand-edited configs.
4. Run `lazyai-cli update --check`, then `lazyai-cli update` when ready.
5. Run `lazyai-cli doctor` to verify generated files and drift.
