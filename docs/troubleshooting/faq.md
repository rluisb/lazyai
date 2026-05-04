# FAQ

## General

### Is ai-setup published to npm?

No. Install and run it directly from GitHub:

```bash
npx github:ricardoborges-teachable/ai-setup init
```

### What is the minimum Node.js version?

Node.js `>=20.12.0` is required. The Go binary itself does not need Node, but the npm bootstrap does.

### Can I use ai-setup without installing anything permanently?

Yes. `npx github:ricardoborges-teachable/ai-setup` downloads and runs the latest version on each invocation without a local install.

### What does `ai-setup doctor` check?

- Whether `.ai-setup.json` exists and is readable
- Whether tracked files exist and match their recorded hashes
- Whether required tool-native directories are present
- Drift between library source and installed skills (with `--skills-check`)
- Migration drift against a clean ai-setup state (with `--migration-check`)

## Setup

### I already have a `.opencode/` or `.claude/` setup. Can I migrate?

Yes. Use `ai-setup import` or `ai-setup migrate`:

```bash
ai-setup import --preview
ai-setup import ../legacy-project --strategy preserve
```

Supported sources: OpenCode, Claude Code, GitHub Copilot.

### What is the difference between `compile` and `update`?

- `compile` regenerates tool-native files from the canonical `.ai/` layer without changing library content.
- `update` refreshes library content (agents, skills, rules, templates) from the bundled source and resolves conflicts.

### Can I stop using ai-setup without losing my files?

Yes:

```bash
ai-setup eject
```

This removes `.ai-setup.json` and stops management, but leaves all generated files in place.

### How do I add a tool after init?

```bash
ai-setup add claude-code
ai-setup compile
```

## Scopes

### Can I change scope after setup?

You must re-run `ai-setup init` with the desired scope. The previous `.ai-setup.json` will be replaced; back it up first if needed.

### Does workspace scope modify my code repos?

No. Referenced repos are scanned for stack detection but are **never** modified. All generated files live in the planning repo.

## Orchestration

### Does enabling orchestration change my existing setup?

No. Orchestration is opt-in. If you never enable `orchestrator`, nothing changes.

### What runs the orchestrator?

The orchestrator is a Go runtime (`ai-setup-orchestrator`) invoked via `connect` as an MCP server. Your host CLI tool (Claude Code, OpenCode, Copilot) remains the execution surface.

### Is A2A remote execution enabled by default?

No. A2A is a config/seam only. The default execution model uses the native host CLI directly.

### Where does orchestration state live?

Runtime state, persistence, and handoff artifacts are managed by the `ai-setup-orchestrator` MCP server, not by `ai-setup` itself.

## MCP

### How do I enable an MCP server after init?

Edit `.ai/mcp.json`, set the server's `enabled` flag to `true`, then run:

```bash
ai-setup compile
```

### Where do I put API keys for MCP servers?

`ai-setup` generates `.env.example` with required variable names, but never writes real secrets. Add actual values to your environment or a local `.env` file (which should be ignored by git).

## Update and conflict

### What happens if a library file conflicts with my custom version?

`ai-setup update` prompts before overwriting customized files. Use `--force` to auto-overwrite with a backup saved under `.ai-setup-backup/`.

### How do I preview updates before applying them?

```bash
ai-setup update --check
ai-setup doctor --skills-check
```

### How do I upgrade the binary itself?

```bash
ai-setup update-self --check
ai-setup update-self
```

## Troubleshooting

### `ai-setup init` fails with "unsupported tool"

Check that the tool ID is one of: `opencode`, `claude-code`, `copilot`. IDs are case-sensitive.

### `ai-setup compile` does not generate files for a tool

Verify the tool is listed in `.ai-setup.json` under `tools`. If missing, run `ai-setup add <tool>` and then `ai-setup compile`.

### `ai-setup doctor` reports missing files after I deleted them

`doctor` tracks the manifest exactly. If you intentionally removed a file, you can re-run `ai-setup update` to recreate it, or run `ai-setup eject` to stop tracking.
