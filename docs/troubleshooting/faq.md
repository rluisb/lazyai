# FAQ

## General

### Is LazyAI published to npm?

No. LazyAI is Go-only. Install commands with `go install`:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

### What is the minimum Go version?

Go 1.26+ is required to install or build the LazyAI commands from source.

### Can I use LazyAI without Node.js?

Yes. LazyAI does not require Node, npm, npx, or pnpm for normal usage.

### What does `lazyai-cli doctor` check?

- Whether `.ai-setup.json` exists and is readable
- Whether tracked files exist and match their recorded hashes
- Whether required tool-native directories are present
- Drift between library source and installed skills (with `--skills-check`)
- Migration drift against a clean LazyAI state (with `--migration-check`)

## Setup

### I already have a `.opencode/` or `.claude/` setup. Can I migrate?

Yes. Use `lazyai-cli import` or `lazyai-cli migrate`:

```bash
lazyai-cli import --preview
lazyai-cli import ../legacy-project --strategy preserve
```

Supported sources: OpenCode, Claude Code, GitHub Copilot.

### What is the difference between `compile` and `update`?

- `compile` regenerates tool-native files from the canonical `.ai/` layer without changing library content.
- `update` refreshes library content (agents, skills, rules, templates) from the bundled source and resolves conflicts.

### Can I stop using LazyAI without losing my files?

Yes:

```bash
lazyai-cli eject
```

This removes `.ai-setup.json` and stops management, but leaves all generated files in place.

### How do I add a tool after init?

```bash
lazyai-cli add claude-code
lazyai-cli compile
```

## Scopes

### Can I change scope after setup?

You must re-run `lazyai-cli init` with the desired scope. The previous `.ai-setup.json` will be replaced; back it up first if needed.

### Does workspace scope modify my code repos?

No. Referenced repos are scanned for stack detection but are **never** modified. All generated files live in the planning repo.

## Retired orchestration runtime

### Where did the old orchestration runtime go?

The dedicated workflow/task runtime was removed from the active product surface. Current setup uses OpenCode, Claude Code, and Copilot adapters with `primary-agent` as the neutral default.

See [Migration: Fortnite / orchestrator removal](../migration/fortnite-orchestrator-removal.md) for replacements and rollback guidance.

### Is A2A remote execution enabled by default?

No. A2A is not part of the active default setup.

### Where does runtime state live now?

`lazyai-cli` owns the active local setup and managed-file state. Session database migration is covered by the runtime refactor plan before V2 schema changes land.

## MCP

### How do I enable an MCP server after init?

Edit `.ai/mcp.json`, set the server's `enabled` flag to `true`, then run:

```bash
lazyai-cli compile
```

### Where do I put API keys for MCP servers?

`lazyai-cli` generates `.env.example` with required variable names, but never writes real secrets. Add actual values to your environment or a local `.env` file (which should be ignored by git).

## Update and conflict

### What happens if a library file conflicts with my custom version?

`lazyai-cli update` prompts before overwriting customized files. Use `--force` to auto-overwrite with a backup saved under `.ai-setup-backup/`.

### How do I preview updates before applying them?

```bash
lazyai-cli update --check
lazyai-cli doctor --skills-check
```

### How do I upgrade the binary itself?

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

## Troubleshooting

### `lazyai-cli init` fails with "unsupported tool"

Check that the tool ID is one of: `opencode`, `claude-code`, `copilot`. IDs are case-sensitive.

### `lazyai-cli compile` does not generate files for a tool

Verify the tool is listed in `.ai-setup.json` under `tools`. If missing, run `lazyai-cli add <tool>` and then `lazyai-cli compile`.

### `lazyai-cli doctor` reports missing files after I deleted them

`doctor` tracks the manifest exactly. If you intentionally removed a file, you can re-run `lazyai-cli update` to recreate it, or run `lazyai-cli eject` to stop tracking.
