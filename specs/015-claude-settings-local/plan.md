# Plan — 015: Claude Code `--local-secrets` (settings.local.json)

**Date:** 2026-04-20
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions locked in §6)

---

## 1. Objective

Add an opt-in `--local-secrets` flag to `ai-setup init`. When set, ai-setup routes the user's MCP catalog (and Claude-settings overrides) into `.claude/settings.local.json` — a per-machine, gitignored file — instead of the committed surfaces (`.mcp.json`, `.claude/settings.json`). Covers use cases N1 (secret env vars), N2 (permission overrides), N3 (experimental hooks).

Exit: users who pass `--local-secrets` get a `.claude/settings.local.json` that holds their full MCP catalog + any settings overrides, with `.gitignore` updated to cover it. No surprise behavior when flag is absent.

---

## 2. Locked decisions (from research §6)

| # | Answer |
|---|---|
| Q1 | **All three use cases** — N1 (secrets), N2 (permissions), N3 (hooks) |
| Q2 | **Option B** — opt-in `--local-secrets` CLI flag |
| Q3 | **Project AND workspace/global** — flag is accepted at all three scopes |
| Q4 | **All servers** move to settings.local.json (not per-server marking) |
| Q5 | **Yes** — auto-update `.gitignore` to cover `.claude/settings.local.json` |
| Q6 | **Deep-merge + backup-on-first-touch** — consistent with settings.json |

### Scope interpretation (noted for review)

- **Project/workspace:** `<target>/.claude/settings.local.json` — standard location, read by Claude CLI's Local scope.
- **Global:** `~/.claude/settings.local.json` — **non-standard**. Claude CLI does not document a user-level local file; user-scope settings live in `~/.claude/settings.json`. We still write the file to give users a clean separation between committed defaults and local overrides, but note in docs that it is ai-setup convention rather than upstream behavior. (If the user prefers global to be a no-op here, flag during human gate.)

---

## 3. Phased breakdown

### Phase 1 — Flag plumbing

**Goal:** Surface the `--local-secrets` bool through `cmd/init.go` → `WizardConfig` → `CompileContext`/`AdapterContext`.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 001 | Add `LocalSecrets bool` to `WizardConfig` (`tui/wizard/wizard.go`) and thread it into `AdapterContext` + `CompileContext` (`internal/adapter/types.go`). Wire through `cmd/init.go` flag registration: `rootCmd.PersistentFlags().BoolVar(&cfg.LocalSecrets, "local-secrets", false, "…")`. | ~60 |

**Exit criteria:** `go build ./...` passes; flag visible under `ai-setup init --help`; no adapter uses the flag yet.

---

### Phase 2 — MCP route split

**Goal:** When `--local-secrets` is set, the Claude Code MCP compile writes to `.claude/settings.local.json` (under `mcpServers`) instead of `.mcp.json`.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 002 | `compileClaudeCodeMCP` branch: if `ctx.LocalSecrets`, call new `writeClaudeSettingsLocal(ctx, servers)`; otherwise keep existing `.mcp.json` path. At global scope with `--local-secrets`: write `~/.claude/settings.local.json` and log a line noting this is ai-setup-convention. | ~70 |
| 003 | `writeClaudeSettingsLocal` — resolves target path per scope, builds `{"mcpServers": toClaudeCodeMcpInner(servers)}` payload, deep-merges via `configmerge.MergeJSONFile` (preserves user-authored keys). Factor out `toClaudeCodeMcpInner` from the current `toClaudeCodeMcp` so both callers reuse server translation. | ~80 |

**Exit criteria:**
- With flag: `.claude/settings.local.json` contains `mcpServers`; `.mcp.json` is NOT emitted (or emitted as empty `{}` stub — see Q below).
- Without flag: byte-identical behavior to today.
- Re-run with pre-existing user keys: those are preserved; `.bak` created on first merge.

**Open question during implementation:** When `--local-secrets` is set, do we still write an empty `.mcp.json` so the project directory advertises "this repo uses MCP, go see settings.local.json"? I lean toward **skipping** `.mcp.json` entirely — less surface, fewer files to maintain. Confirm during implementation if a team expects the stub.

---

### Phase 3 — Settings.json override routing (N2/N3)

**Goal:** Extend the split beyond MCP — any Claude settings merge that the scaffold would have written into `.claude/settings.json` goes to `.claude/settings.local.json` when the flag is set. This covers permissions, hooks, and other settings overrides.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 004 | Audit `internal/scaffold/` for call sites that write `.claude/settings.json` (beyond MCP). Identify whether each call should be redirected under `--local-secrets`. Likely candidates: `settings.json` permissions merge (if any), hook registration (if any). If no such call-sites exist today (MCP is the only compile output touching this file), **skip the task** and note that N2/N3 are enabled purely by the gitignore + user-authored contents. | ~40 (often 0) |
| 005 | If any call-sites exist: centralize routing in a helper `resolveClaudeSettingsPath(ctx) → path` that returns either `settings.json` or `settings.local.json` based on `ctx.LocalSecrets`. Update all call sites to use it. | ~60 (conditional on 004) |

**Exit criteria:** N2/N3 are at minimum enabled via gitignore + user free-form edits; no regression to settings.json behavior when flag is off.

---

### Phase 4 — Gitignore auto-update

**Goal:** Ensure `.claude/settings.local.json` is listed in the repo's `.gitignore` when ai-setup writes it first (Claude CLI does this on its own writes, but ai-setup may beat it).

| Task | Deliverable | Rough LOC |
|---|---|---|
| 006 | Add `.claude/settings.local.json` to the gitignore suggestion printed by the scaffold (look for the existing "💡 Consider creating a .gitignore" block in `internal/scaffold/` — append the new line there). If ai-setup actually writes to a project scope, append to `.gitignore` if it exists; skip if not (print suggestion). Match existing logic for `.ai/memory/`. | ~50 |

**Exit criteria:** Running `ai-setup init --local-secrets` in a repo with a `.gitignore` appends the new line (idempotent). Running without the flag touches nothing.

---

### Phase 5 — Tests + docs

**Goal:** Cover the new path with unit + integration tests; update knowledge map.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 007 | Unit: `writeClaudeSettingsLocal` — stdio server, sse server, env preserved, deep-merge preserves user keys. | ~80 |
| 008 | Scope-parity: test that `--local-secrets` at project/workspace/global produces the expected file at the expected path; test that it is idempotent. | ~80 |
| 009 | Integration: pre-existing `.claude/settings.local.json` with a user-authored `env` key → re-run with `--local-secrets` → assert user key preserved, ai-setup keys written, `.bak` created. | ~60 |
| 010 | Update `specs/KNOWLEDGE_MAP.md` — mark item complete; add spec 015 entry; note the flag under Packages Reference. | ~10 |

**Exit criteria:** `go test ./... -count=1` green; knowledge map reflects ship; flag documented in `cmd/init.go` help string.

---

## 4. Non-goals (for clarity)

- **Not splitting per-server secret-only routing** — entire catalog moves when the flag is on (locked Q4).
- **Not creating managed-settings.json support** — MDM/managed scope stays out of spec.
- **Not modifying Claude CLI's own `settings.local.json` writes** — ai-setup and the CLI will both deep-merge; backup protects user data on either side.
- **Not changing behavior when the flag is absent** — strict opt-in.

---

## 5. Risk register

| # | Risk | Mitigation |
|---|---|---|
| R1 | User flag-toggles `--local-secrets` between runs; data split inconsistently | Document the behavior in init help string; deep-merge + backup ensure no silent loss. |
| R2 | Global scope write to `~/.claude/settings.local.json` may not be read by Claude CLI | Log a one-line info message at write time; document as ai-setup convention. |
| R3 | `.gitignore` append duplicates the line on repeated runs | Grep for existing line before appending (existing `.ai/memory/` pattern). |
| R4 | Tests run on dev machines with real Claude installs | Use `t.TempDir()` + injected `HomeDir`; avoid real `LookupClaudeBinary`. |

---

## 6. Sequencing & sizing

- Phase 1 — 1 task, ~60 LOC. Flag plumbing.
- Phase 2 — 2 tasks, ~150 LOC. MCP route split (core feature).
- Phase 3 — 0–2 tasks, ~0–100 LOC. Conditional on call-site audit.
- Phase 4 — 1 task, ~50 LOC. Gitignore.
- Phase 5 — 4 tasks, ~230 LOC. Tests + docs.

Total: 8–10 tasks, ~490–590 LOC. Fits within a single-session implementation.

---

## 7. Out of scope / deferred

- Claude plugin manifest (already deferred to its own follow-up).
- Managed / MDM settings handling.
- CLAUDE.local.md (memory analog) — separate artifact, not covered here.
