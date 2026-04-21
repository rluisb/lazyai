# Requirements Checklist — 011: OpenCode Deep Setup

Verification criteria derived from [plan.md](../plan.md). Each AC must be green before the spec is marked complete in `KNOWLEDGE_MAP.md`.

## Acceptance Criteria

### AC-1 — Single canonical config file
- [ ] Install at **project** scope produces exactly one `<project>/.opencode/opencode.jsonc`.
- [ ] Install at **workspace** scope produces exactly one `<workspace>/.opencode/opencode.jsonc`.
- [ ] Install at **global** scope produces exactly one `~/.config/opencode/opencode.jsonc`.
- [ ] No `opencode.json` file remains at any scope after install.
- [ ] If a pre-existing `opencode.json` existed, a `.bak` sidecar preserves its contents.

### AC-2 — `instructions` key resolution
- [ ] Each scope's `opencode.jsonc` contains `"instructions": ["AGENTS.md"]`.
- [ ] The referenced `AGENTS.md` file exists inside the same `.opencode/` root.
- [ ] `opencode debug config` (when binary present) shows `instructions` resolved without errors.

### AC-3 — Agent frontmatter schema conformance
- [ ] Every installed agent file (`<root>/agents/<name>.md`) starts with a YAML frontmatter block.
- [ ] Frontmatter contains at minimum: `description`, `mode`, `tools`, `model`.
- [ ] `mode` ∈ `{primary, subagent, all}`.
- [ ] `tools` matches opencode's accepted shape (map or comma string; confirmed against v1.4.9).
- [ ] `opencode debug agent <name>` (when binary present) returns zero errors for each agent.

### AC-4 — MCP deep-merge preserves user servers
- [ ] Test: pre-seed `opencode.jsonc` with a hand-authored MCP server entry → run `ai-setup compile` → user entry still present.
- [ ] Test: ai-setup-managed server toggled disabled → re-compile → user entries unchanged, managed server updated.
- [ ] Test: a brand-new compile onto an empty config produces identical output to the prior overwrite behavior (no regression).

### AC-5 — Commands and Modes
- [ ] `library/opencode/commands/*.md` bundle is embedded and installable.
- [ ] `library/opencode/modes/*.md` bundle is embedded and installable.
- [ ] Installed commands land at `<root>/commands/<name>.md` at all 3 scopes.
- [ ] Installed modes land at `<root>/modes/<name>.md` at all 3 scopes.
- [ ] Wizard offers per-asset selection for opencode commands and modes (custom preset only, matching spec 010 pattern).
- [ ] Store persists `OpenCodeCommands` and `OpenCodeModes` selection slices.

### AC-6 — Post-install validation (opt-in)
- [ ] When `opencode` is on `PATH` and opencode is selected, `ai-setup init` runs `opencode debug config` and `opencode debug agent <name>` after install.
- [ ] Zero warnings on a clean, fresh install at each scope.
- [ ] Warnings are non-fatal (init exits 0) and printed via the existing logger.
- [ ] When `opencode` is **not** on `PATH`, no validation runs and no error is emitted.

### AC-7 — Plugin install flow (opt-in)
- [ ] Wizard offers an "OpenCode plugins" multi-select step, gated on (a) opencode selected, (b) binary on PATH.
- [ ] Store persists `OpenCodePlugins` selection.
- [ ] For each selected plugin at **global** scope: `opencode plugin <module> -g` is invoked.
- [ ] For each selected plugin at **project/workspace** scope: `opencode plugin <module>` is invoked with `cwd` = target dir.
- [ ] Plugin install failures are surfaced as non-fatal warnings; init exits 0.
- [ ] When `opencode` is not on `PATH`, no plugin step is shown and no shell-out occurs.

## Non-Regression

- [ ] `go test ./... -count=1` is green.
- [ ] `go vet ./...` is green.
- [ ] Manual smoke: `go run . init` in a temp dir with opencode selected produces the expected layout.
- [ ] Scope-parity tests for opencode cover install + compile + validation across all 3 scopes.
- [ ] Other adapters (Claude, Gemini, Codex, Copilot) are untouched.

## Knowledge Updates (before closing the spec)

- [ ] `specs/KNOWLEDGE_MAP.md`: spec 011 row added with status.
- [ ] `CLAUDE.md`: Codebase Map updated if new packages/dirs introduced.
- [ ] ADR drafted in `specs/adrs/` if any architectural decision needs recording (e.g., MCP managed/user-authored detection strategy).
