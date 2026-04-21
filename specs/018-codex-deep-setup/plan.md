# Plan — 018: Codex Deep Setup

**Date:** 2026-04-21
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions locked in §5)

---

## 1. Objective

Close the Codex parity gap with specs 011/012/013/017 within the constraints of what Codex upstream actually supports. Three deliverables:
1. **Fix broken `RunHeadlessValidation`** — add `--skip-git-repo-check` so the probe stops silently failing.
2. **Create `library/codex/`** as a per-tool asset home; ship a Codex-specific AGENTS.override starter template.
3. **Add a post-install MCP summary** via `codex mcp list` so users get the same courtesy signal Claude (spec 012) already has.

Exit: `codex exec` probe succeeds (or emits a clean, actionable message); `ai-setup init --tools codex` emits a summary line indicating how many MCP servers are registered.

---

## 2. Locked decisions (from research §5)

| # | Answer |
|---|---|
| Q1 | **Option C** — validation fix + `library/codex/` + post-install MCP summary |
| Q2 | **Raw template** — ship `library/codex/AGENTS.override.template.md` with `[YOUR_*]` placeholders for the recipient to fill (consistent with `GEMINI.template.md` and the Gemini extension's raw template choice) |
| Q3 | **Always run** `codex mcp list` summary when the `codex` binary is present (non-fatal on failure, matches Claude's behavior) |
| Q4 | N/A (not shipping Codex-specific skills this spec) |
| Q5 | **Leave generic `library/skills/` shared** across all tools — no tool-specific resolver needed here |

---

## 3. Phased breakdown

### Phase 1 — Fix broken validation

**Goal:** Make `codex exec` actually succeed (or fail cleanly for a documented reason) when run from ai-setup's install flow.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 001 | `CodexAdapter.RunHeadlessValidation` — add `--skip-git-repo-check` to the argv. Keep the 30s timeout and non-fatal error handling. Update the log line to include the full command string so users see exactly what ran. | ~15 |

**Exit criteria:** Running `ai-setup init --tools codex` in a `t.TempDir()` workspace (non-git, non-trusted) no longer emits "Not inside a trusted directory"; test harness validates the flag is present in the argv.

---

### Phase 2 — `library/codex/` per-tool dir + AGENTS.override starter

**Goal:** Give Codex the per-tool layout parity of specs 011/012/013/017. Ship one starter asset (AGENTS.override template) so the dir isn't empty.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 002 | Create `library/codex/AGENTS.override.template.md` — raw template with `[YOUR_ORG]`, `[YOUR_TEAM]`, `[YOUR_PROJECT_DESCRIPTION]`, and a Codex-idiom comment block at the top ("This file is loaded hierarchically by `codex`; see https://developers.openai.com/codex/guides/agents-md."). Body mirrors `GEMINI.template.md` structure for team familiarity. | ~90 (content only) |
| 003 | `library.CodexAssetsDir() string` helper returning `filepath.Join(Root(), "codex")`. Add the constant `CodexSubdir = "codex"` for embedded-FS readers. Follows the pattern of `CopilotAgentsDir` / `GeminiCommandsDir`. | ~15 |
| 004 | Codex adapter wiring — when `library/codex/AGENTS.override.template.md` exists in the library FS and the target's `.codex/AGENTS.override.md` (project/workspace) or `~/.codex/AGENTS.override.md` (global) does NOT exist, copy it verbatim. Skip if the destination already exists (user-authored content wins). No placeholder fill — user fills manually (locked Q2). | ~60 |

**Exit criteria:** `library/codex/AGENTS.override.template.md` exists and is embedded; re-running install in a fresh temp dir produces `.codex/AGENTS.override.md`; existing override files are never overwritten.

---

### Phase 3 — Post-install MCP summary

**Goal:** When `codex` is on PATH, run `codex mcp list` and log a single-line summary of how many servers are registered, matching the Claude Code post-install summary pattern from spec 012.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 005 | `displayCodexInstallSummary(ctx, configRoot, isGlobal)` — mirrors `displayInstallSummary` from `claudecode.go`. Runs `codex mcp list --json` (pin the flag; parse stdout; count entries). If `--json` isn't supported by the installed Codex version, fall back to counting non-empty lines in plaintext output. Non-fatal on any error. Called at the end of `CodexAdapter.Install` after all file writes. | ~90 |
| 006 | Use the existing `codex mcp list` invocation pattern (same `exec.CommandContext` + timeout + CombinedOutput shape used in `installCodexMCPViaCLI`). Factor a small helper `runCodexCommand(ctx context.Context, workingDir string, args ...string) ([]byte, error)` if duplication gets annoying. | ~30 |

**Exit criteria:** After `ai-setup init --tools codex` completes, stderr contains a line like `[codex] Install summary (scope: project) • 2 MCP server(s) registered`. Absent-binary case logs a courtesy "codex CLI not on PATH" and moves on.

---

### Phase 4 — Tests + docs

**Goal:** Golden coverage of the three new paths.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 007 | Unit test for `RunHeadlessValidation`: use an injectable `exec.Command`-like shim (or skip the real exec and just inspect the built argv). Assert `--skip-git-repo-check` is present and the prompt arg is second. | ~60 |
| 008 | Adapter test: install Codex into `t.TempDir()` with a test FS that includes `library/codex/AGENTS.override.template.md`. Assert `.codex/AGENTS.override.md` is written and matches the source bytes. Idempotency test: pre-seed a user-authored `AGENTS.override.md`, run install, assert it's NOT overwritten. | ~80 |
| 009 | Post-install summary test: stub the MCP list runner (inject a runner interface if one doesn't exist; otherwise use a build-tag-guarded fake). Assert the summary line contains "MCP server(s) registered" and a numeric count derived from the stub output. | ~70 |
| 010 | Update `specs/KNOWLEDGE_MAP.md`: mark spec 018 complete; add Packages Reference rows for `library.CodexAssetsDir` and the summary helper. | ~10 |

**Exit criteria:** `go test ./... -count=1` green; knowledge map reflects ship; `ai-setup init --tools codex` prints the expected summary when `codex` is on PATH.

---

## 4. Non-goals

- **No extension/plugin generator** — Codex has no upstream bundle format (hard constraint from research §3 G3).
- **No custom slash commands** — upstream doesn't support user-defined ones; skills are the extensibility surface and are already shipped from `library/skills/`.
- **No Codex-specific skills** — no concrete demand; follow-up if requested.
- **No `~/.codex/prompts/` emission** — deprecated upstream; don't ship onto a deprecated surface.
- **No TOML schema changes** — config.toml deep-merge already handles everything. Keep the merge payload identical.

---

## 5. Risk register

| # | Risk | Mitigation |
|---|---|---|
| R1 | `--skip-git-repo-check` flag rename/removal in future Codex versions | Non-fatal error handling; log full argv so a regression surfaces in user output; no hard dependency on the probe succeeding. |
| R2 | `codex mcp list --json` not supported by the installed Codex version | Try `--json` first; if exit != 0 or output isn't valid JSON, fall back to line counting. Never fail install on a summary parse error. |
| R3 | Test harness for exec-calling code is fragile | Test the argv construction, not the exec outcome (same pattern spec 012 used for Claude probes). No real `codex` binary call in unit tests. |
| R4 | AGENTS.override template doc link rot | The doc URL is in a comment, not a test assertion. If the link moves, update the template in a follow-up commit; no test break. |
| R5 | Embed FS doesn't include `library/codex/` because `go:embed all:library` skips empty dirs | Shipping `AGENTS.override.template.md` inside the dir guarantees it's embedded. Unit test confirms the file is readable via `library.GetLibraryFS()`. |

---

## 6. Sequencing & sizing

| Phase | Tasks | LOC |
|---|---|---|
| 1. Validation fix | 001 | ~15 |
| 2. Library dir + AGENTS.override | 002–004 | ~165 |
| 3. Post-install MCP summary | 005–006 | ~120 |
| 4. Tests + docs | 007–010 | ~220 |
| **Total** | **10 tasks** | **~520 LOC** |

Fits a single implementation session; each phase leaves the tree buildable and tested.

---

## 7. Follow-ups queued for later specs

- Codex-specific skills under `library/codex/skills/` with `agents/openai.yaml` sidecars (G4 from research).
- Delete `library/commands/` fallback path once one release ships (queued from spec 017; tracked here for convenience).
- Snapshot tests for library assets + compiled output (still pending from spec 009 research).
