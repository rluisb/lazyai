# Plan — 014: Copilot Global MCP Compile (`mcp-config.json`)

**Date:** 2026-04-20
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions locked in §6)

---

## 1. Objective

Close the "Copilot `--drive-cli`" deferred item by emitting `~/.copilot/mcp-config.json` at **all scopes** when the Copilot CLI probe passes, so the standalone `@github/copilot` CLI receives MCP servers from ai-setup. Preserve the existing `.vscode/mcp.json` emission for the VS Code extension at project/workspace.

> **Name correction:** spec officially renamed from "Copilot --drive-cli" to **"Copilot global MCP compile (mcp-config.json)"** per locked Q1.

Exit: the Copilot CLI at `~/.copilot/` sees the same MCP catalog as the VS Code extension at `.vscode/mcp.json`, with shapes correctly translated.

---

## 2. Locked decisions (from research §6)

| # | Answer |
|---|---|
| Q1 | **Rename** to "global MCP compile (mcp-config.json)" |
| Q2 | **Option A** — direct-write with deep-merge |
| Q3 | **Emit at project/workspace too** (write to `~/.copilot/mcp-config.json` even when scope is project/workspace, so the CLI works regardless of install mode) |
| Q4 | **Abstract** the schema transforms behind a small interface |
| Q5 | **Probe gating still applies** (only emit if `copilot` on PATH or `~/.copilot/` exists) |

---

## 3. Phased breakdown

### Phase 1 — Abstract the Copilot MCP transform

**Goal:** Introduce a transform abstraction so we can emit two different schemas (VS Code `servers` / CLI `mcpServers`) without duplicating server translation logic.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 001 | Refactor `toCopilotMcp` into a shared `toCopilotServerEntries(servers)` returning a `map[string]any` of stdio/sse entries (the per-server shape is identical across both surfaces). Add two callers: `toCopilotVSCodeMcp(servers)` → `{servers, inputs?}` (current) and `toCopilotCLIMcp(servers)` → `{mcpServers}`. | ~90 |

**Exit criteria:** `go test ./... -count=1` green. Existing tests for `.vscode/mcp.json` still pass with same output bytes.

---

### Phase 2 — Global emitter with deep-merge

**Goal:** Write `~/.copilot/mcp-config.json` at global scope using `configmerge.MergeJSONFile` (preserves user-authored keys; backup-on-first-touch).

| Task | Deliverable | Rough LOC |
|---|---|---|
| 002 | `compileCopilotCLIMcp(ctx)` — resolves `~/.copilot/mcp-config.json` via `ctx.HomeDir`, calls probe check (reject if no CLI + no `~/.copilot/`), runs `toCopilotCLIMcp`, deep-merges via `configmerge.MergeJSONFile`, appends tracked file record. | ~80 |
| 003 | Wire `compileCopilotCLIMcp` into `CompileMCPForTool`. At project/workspace scope: call **both** `compileCopilotMCP` (VS Code) and `compileCopilotCLIMcp` (CLI). At global scope: call `compileCopilotCLIMcp` only. Remove the global-scope skip in the existing probe block (it still skips when probe fails, but only within the new helper). | ~40 |

**Exit criteria:**
- At project/workspace with probe passing: both `.vscode/mcp.json` and `~/.copilot/mcp-config.json` written.
- At global with probe passing: only `~/.copilot/mcp-config.json` written.
- At any scope with probe failing: no CLI file written; VS Code file still written at project/workspace (that is an extension-level artifact, not a CLI one).
- `~/.copilot/mcp-config.json.bak` created on first overwrite of a user-authored file.

---

### Phase 3 — Tests, scope-parity, docs

**Goal:** Cover the new emission paths with unit + scope-parity tests; update knowledge map.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 004 | Unit tests for `toCopilotCLIMcp`: stdio entry, sse entry, env preservation, absence of `inputs` key. Golden JSON assertion. | ~80 |
| 005 | Scope-parity addition: extend `TestCompileMCPForTool_ScopeParity` / similar to assert `~/.copilot/mcp-config.json` is written when probe passes (use the same probe-passing pattern from `TestCopilotAdapter_GlobalScope_Emits`). | ~60 |
| 006 | Integration test — write a pre-existing `~/.copilot/mcp-config.json` with a user key, re-run compile, assert user key preserved and `.bak` created. | ~50 |
| 007 | Update `specs/KNOWLEDGE_MAP.md` — mark item complete; add entry for spec 014; note the two-shape split in the Packages Reference section. | ~10 |

**Exit criteria:** `go test ./... -count=1` green; knowledge map reflects ship.

---

## 4. Non-goals (for clarity)

- **Not implementing `copilot mcp add`** — upstream has no such command.
- **Not modifying the `.vscode/mcp.json` output** — tests for it stay unchanged (byte-identical).
- **Not writing `~/.copilot/` structure beyond `mcp-config.json`** — agents/instructions are already handled by spec 013 install path.
- **Not registering Copilot at runtime via `--additional-mcp-config` flag** — that's a user-runtime choice; ai-setup stops at file placement.

---

## 5. Risk register

| # | Risk | Mitigation |
|---|---|---|
| R1 | User-authored `mcp-config.json` accidentally clobbered | `configmerge.MergeJSONFile` + backup-on-first-touch (existing pattern). |
| R2 | Schema drift between VS Code (`servers`) and CLI (`mcpServers`) | Both transforms share server-entry builder; a unit test asserts per-server shape identical. |
| R3 | Test flakes if host machine has `~/.copilot/` or `copilot` on PATH | Use `t.TempDir()` for `HomeDir`, avoid real `LookupCopilotBinary` — inject probe result via context or env. Follow `TestCopilotAdapter_GlobalScope_Emits` pattern. |
| R4 | MCP `inputs` array accidentally leaks into CLI output (VS-Code-only concept) | Separate builder functions; `toCopilotCLIMcp` has no inputs branch. |

---

## 6. Sequencing & sizing

- Phase 1 — 1 task, ~90 LOC. Refactor + test-stable.
- Phase 2 — 2 tasks, ~120 LOC. Core feature.
- Phase 3 — 4 tasks, ~200 LOC. Tests + docs.

Total: 7 tasks, ~410 LOC. Fits within a single-session implementation.

---

## 7. Out of scope / deferred

- OpenCode `--drive-cli` — permanently deferred (interactive-only).
- Copilot cloud/marketplace publishing — parked.
- Runtime validation via `copilot --check` or similar — no such command exists.
