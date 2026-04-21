# Plan — 012: Claude Code Deep Setup (Global / Project / Workspace)

**Date:** 2026-04-19
**Author:** Ricardo (with Scout + claude-code-guide agents)
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions in §8)

---

## 1. Objective

Deliver a Claude Code install path that is **layout-conformant at all three scopes**, fixes the 12 defects catalogued in research §4, adds starter `commands/` + `output-styles/` assets on parity with OpenCode, and **opportunistically delegates MCP registration to the `claude` CLI** when the binary is on PATH (with a transparent direct-write fallback).

Each phase below leaves the codebase in a shippable state. Tests pass (`go test ./... -count=1`) at the end of every phase.

---

## 2. Phased breakdown

### Phase 1 — Structural fixes (direct-write only)

**Goal:** Close the known defects without introducing any CLI dependency. Keeps blast radius small and gives a clean baseline for later phases.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 001 | Global scope — move agents into `~/.claude/agents/<name>.md` (remove flat-at-root layout); move tool-context from `~/.claude/CLAUDE.md` to `~/.claude/agents/CLAUDE.md`; leave personal `~/.claude/CLAUDE.md` alone on re-run (first-install-only template fill) | ~80 |
| 002 | Orchestrator agent at global scope — remove `!isGlobal` gate at `claudecode.go:95` so orchestrator follows `IsOrchestratorEnabled` everywhere | ~15 |
| 003 | Source `rules/typescript.md` from `library/rules/` instead of the hardcoded string in `claudecode.go:63` (create the library asset if missing) | ~30 |
| 004 | Fix agent `tools` frontmatter delimiter to whitespace (Claude's canonical form) for Claude only; OpenCode keeps its existing delimiter | ~25 |

**Exit criteria:** `go test ./... -count=1` passes. Manual smoke: `go run . init` against a temp project + a throwaway `$HOME` produces the correct layout at all three scopes.

---

### Phase 2 — Starter library assets for `commands/` + `output-styles/`

**Goal:** Ship Claude-targeted starter content on parity with OpenCode so Claude scope installs have immediate value.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 005 | `library/claudecode/commands/{review,test,commit}.md` with Claude frontmatter (`description`, `argument-hint`, `allowed-tools`) | ~60 (mostly content) |
| 006 | `library/claudecode/output-styles/{terse,explanatory}.md` with Claude frontmatter (`name`, `description`, `keep-coding-instructions`) | ~40 (mostly content) |
| 007 | Adapter wiring — walk `library/claudecode/commands/` and `output-styles/`, copy into `<scope-root>/commands/` and `<scope-root>/output-styles/` at all three scopes; add the directories to scope-parity test expectations | ~60 |

**Exit criteria:** New directories exist and are populated at every scope; scope-parity tests updated and green.

---

### Phase 3 — CLI orchestration for MCP (opt-in, silent fallback)

**Goal:** When `claude` is on PATH, delegate MCP server registration to `claude mcp add` / `claude mcp add-json` so Claude owns its on-disk format. When absent, fall back to the existing direct-write path.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 008 | `internal/adapter/claude_cli.go` — probe helper (`LookupClaudeBinary`) and `ClaudeCLIRunner` interface with an injectable default implementation (matches the `CmdRunner` pattern used in `opencode_validate.go`) | ~90 |
| 009 | MCP install via CLI — in `claudecode.go`, when CLI is available, loop enabled MCP servers and invoke `claude mcp add-json <name> <json>` with `-s user\|project` based on scope; check presence first via `claude mcp get <name>` to avoid duplicates; emit a single-line warning and fall back on error or missing binary | ~120 |
| 010 | MCP compile via CLI at project scope — make `compileClaudeCodeMCP` CLI-aware too, so subsequent compiles reconcile against what `mcp list` reports (direct `.mcp.json` write remains the fallback) | ~70 |

**Exit criteria:** With `claude` present, `.mcp.json` (project) and `~/.claude/settings.json` `mcpServers` (user) are produced by the CLI; `claude mcp list` enumerates everything we registered. With `claude` absent, behavior matches end-of-Phase-1.

---

### Phase 4 — Tests, verification, docs

**Goal:** Lock the fixes in with regression tests and ship doc updates.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 011 | Frontmatter schema tests — parse every emitted agent, skill, command, output-style file and assert required fields + delimiter correctness for Claude | ~120 |
| 012 | Scope-parity regression test — explicit assertion that global scope places agents under `agents/` subdir (catches the research-§4.3 defect on any future regression); orchestrator presence at global when enabled | ~80 |
| 013 | CLI-integration test for MCP — inject a fake `ClaudeCLIRunner` that records calls; assert correct `mcp add-json` invocation shape per scope; assert fallback path triggers when probe returns "not found" | ~100 |
| 014 | Post-install summary — when CLI is present, run `claude mcp list` and `claude agents --setting-sources user` after install and include a one-block summary in the install output; failure is non-fatal | ~50 |
| 015 | Docs — update `specs/KNOWLEDGE_MAP.md` with spec 012 row, add `settings.local.json` to Pending / Follow-up, append package additions to root `CLAUDE.md` codebase map if any | ~30 |

**Exit criteria:** All new tests green; `go vet ./...` clean; KNOWLEDGE_MAP reflects the completed spec.

---

## 3. Scope hygiene

**In scope**
- Every defect enumerated in research §4.3 that pertains to Claude Code.
- `commands/`, `output-styles/` starter content + adapter wiring.
- CLI orchestration for MCP at all three scopes, with direct-write fallback.
- Test coverage for layout correctness, frontmatter contracts, and CLI call shape.

**Explicitly out of scope (per decisions §8)**
- `.claude/settings.local.json` (Q5 — follow-up ticket).
- Authoring a `.claude-plugin/plugin.json` manifest (Q6 — later spec).
- Any workspace-only Claude branches (Q7 — workspace = project-identical).
- Other tools (OpenCode, Gemini, Codex, Copilot) — untouched.

**Explicit non-goals**
- No ADR required. None of the decisions cross an architectural boundary large enough to warrant one (MCP-via-CLI is a localized adapter change, not a new system boundary). If Phase 3 surfaces something unexpected we'll promote it to an ADR at that checkpoint.

---

## 4. Risk register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `claude mcp add-json` payload schema diverges from our canonical MCP shape | Medium | High | Task 009 ships a small transformer + unit test fixtures pinning the payload shape; version check in probe if upstream adds versioning. |
| Running `claude mcp add` twice for the same server errors out | High | Medium | Task 009: pre-check via `claude mcp get <name>`; skip if present, remove-then-add if force-refresh requested. |
| Silent-fallback path hides real CLI failures from users | Medium | Medium | Task 009: emit a single warning line with the reason (binary-missing vs invocation-error); include in the install summary. |
| Moving agents at global scope breaks existing user installs | Medium | High | Task 001: on install, detect flat-layout agent files at `~/.claude/*.md` and migrate them to `agents/` (or at least warn + document). |
| `NormalizeToolsFrontmatter` change leaks into OpenCode path | Low | Medium | Task 004: change is Claude-scoped via a new parameter/helper; existing OpenCode call site keeps comma delimiter; explicit test. |
| CI doesn't have `claude` binary — Phase 3 tests can't run the real CLI | High | Low | Tests use an injected `ClaudeCLIRunner` fake; no real CLI required in CI. Real-CLI smoke is manual per spec `## Verification`. |

---

## 5. Verification rounds (per CLAUDE.md protocol)

This is a **moderate** task → **2 verification rounds** per phase:

1. Requirements re-check against `checklists/requirements.md` at end of each phase.
2. Edge-case pass: re-run with CLI absent, re-run on top of an existing install, re-run at all three scopes.

Integration-boundary round is folded into Phase 4 by the test tasks themselves.

---

## 6. Open items that could reshape the plan

None blocking — the decisions in research §8 cover every branch point. If during Phase 3 the `claude mcp add-json` schema turns out to not accept our canonical MCP shape, I'll escalate before writing the fallback path.

---

## 7. Task files

See `tasks/001-*.md` through `tasks/015-*.md`. Each file is self-contained: goal, files to touch, acceptance criteria, test plan.
