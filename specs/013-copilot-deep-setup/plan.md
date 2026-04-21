# Plan — 013: GitHub Copilot Deep Setup (VS Code + Standalone CLI)

**Date:** 2026-04-20
**Author:** Ricardo (with Scout agent)
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions locked in §10)

---

## 1. Objective

Deliver a Copilot install path that is **layout-conformant for both surfaces** (VS Code extension + standalone `@github/copilot` CLI), lifts the Copilot × global block, closes the 11 defects catalogued in research §4.3, and ships starter library assets (`agents/`, `instructions/`) on parity with OpenCode (spec 011) and Claude (spec 012).

Each phase below leaves the codebase in a shippable state. Tests pass (`go test ./... -count=1`) at the end of every phase.

---

## 2. Phased breakdown

### Phase 1 — Library assets (content only, no behavior change)

**Goal:** Ship the raw library content so later phases have something to copy.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 001 | `library/copilot/agents/{planner,builder,scout,reviewer}.agent.yaml` using the zod-confirmed schema from research §2.2; orchestrator gated on `EnableServers` via wiring in later phase | ~160 (mostly content) |
| 002 | `library/copilot/instructions/{typescript,go,tests}.instructions.md` with `applyTo` globs (TS `**/*.{ts,tsx}`, Go `**/*.go`, tests `**/*_test.*,**/*.test.*,**/*.spec.*`) | ~90 (content) |
| 003 | Embed new dirs in `internal/library/library.go` and add a schema unit test (parse every `*.agent.yaml` + every `*.instructions.md`, assert required fields, assert `applyTo` non-empty for instructions) | ~80 |

**Exit criteria:** `go test ./... -count=1` green. New assets listed by `internal/library/integration_test.go`.

---

### Phase 2 — Project/Workspace: emit `agents/` + `instructions/` + MCP inputs

**Goal:** VS-Code-surface parity today, independent of any global-scope work. Closes G2, G3, G5, G7 at project/workspace scope.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 004 | Adapter wiring — `CopilotAdapter.Install` walks `library/copilot/agents/` → `.github/agents/<name>.agent.yaml` and `library/copilot/instructions/` → `.github/instructions/<name>.instructions.md`. Honors selection sets. | ~100 |
| 005 | Skills migration — replace the current `skills → *.prompt.md` transform with `skills → .github/agents/<skill>.agent.yaml`. Drop `EnsureModeAgentFrontmatter` Copilot call-site. `library/prompts/` continues to be emitted as `.prompt.md`. | ~80 |
| 006 | `.vscode/mcp.json` `inputs` scaffolding — update `toCopilotMcp` (`mcp_compiler.go:339`) to scan `env` values for `${VAR}`-style placeholders; for each unique placeholder emit a matching `inputs: [{type:"promptString",id:"VAR",password:true}]` entry. Omit key when no placeholders found. | ~70 |

**Exit criteria:** `.github/agents/`, `.github/instructions/` populated at project + workspace scope; `.vscode/mcp.json` emits `inputs` when placeholders present and omits it otherwise. Scope-parity tests updated.

---

### Phase 3 — Lift Copilot × global block + probe

**Goal:** Unblock global scope, gated on the `copilot` CLI / `~/.copilot/` dir being present.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 007 | `internal/adapter/copilot_cli.go` — `LookupCopilotBinary()` (PATH search) + `CopilotCLIRunner` interface, mirroring `claude_cli.go` from spec 012. Additionally a `CopilotHomePresent(homeDir)` helper that checks for `~/.copilot/`. | ~90 |
| 008 | Flip scope gating — `scope.go:28` removes the Copilot × global short-circuit; `globalpaths.go:73-82` returns `true` for Copilot and resolves `~/.copilot`. Adapter emits a one-line warning + no-op when both binary and `~/.copilot/` are absent at global scope. | ~60 |

**Exit criteria:** `IsScopeSupported(ToolIdCopilot, SetupScopeGlobal)` returns true; adapter silent no-ops when probe fails; scope-parity matrix includes (copilot, global).

---

### Phase 4 — Global-scope emitters

**Goal:** Write the actual files at `~/.copilot/` when probe passes.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 009 | Global agents + instructions — emit `~/.copilot/agents/*.agent.yaml` (library + migrated skills, same content as `.github/agents/` at project scope) and `~/.copilot/copilot-instructions.md` (template-filled first-install only; re-run leaves it alone, mirrors spec 010 CLAUDE.md hybrid fill) | ~110 |
| 010 | MCP user-scope compile — extend `compileCopilotMCP` to take scope into account; at global scope, deep-merge `~/.copilot/mcp-config.json` via `configmerge.MergeJSONFile` with backup-on-first-touch. Managed servers win on key collision; user-authored preserved. Schema per research §2.4 (`mcpServers` top-level, per-server `type`/`command`/`args`/`env` for stdio, `type`/`url`/`headers` for http/sse). | ~130 |

**Exit criteria:** At global scope with probe passing, `~/.copilot/agents/`, `~/.copilot/copilot-instructions.md`, `~/.copilot/mcp-config.json` all produced. User-authored `mcp-config.json` keys preserved across re-run.

---

### Phase 5 — Post-install validation

**Goal:** Smoke-test each shipped agent via the `copilot` CLI when present.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 011 | `CopilotAdapter.RunHeadlessValidation` — for each emitted `.agent.yaml` (at whichever scope the install targeted), if `copilot` is on PATH, run `copilot --agent <name> -p "ai-setup validation ping" --allow-all-tools -s` with a short timeout (5s). Non-zero exit = single-line warning to stderr; never fatal. Injectable runner for tests. | ~100 |

**Exit criteria:** `CanRunHeadless` returns true when probe passes; tests inject a fake runner that records invocations and a fake that returns non-zero, assert both paths.

---

### Phase 6 — Tests + docs

**Goal:** Lock everything in with regression tests and ship documentation updates.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 012 | Scope-parity regression test — explicit (copilot, global) scope matrix entry asserting `~/.copilot/agents/`, `~/.copilot/mcp-config.json`, `~/.copilot/copilot-instructions.md` all present when probe passes; no writes when probe fails. | ~90 |
| 013 | Frontmatter + schema tests — parse every emitted `.agent.yaml` (name, model, prompt required), every `.instructions.md` (applyTo present), every `.chatmode.md` (description required). Unit test for `mcp-config.json` shape at user scope. | ~120 |
| 014 | MCP round-trip + inputs scaffolding tests — seed a user-authored server in `~/.copilot/mcp-config.json`, run compile, assert managed + user servers both present. Seed `.vscode/mcp.json` MCP server with `env:{FOO:"${FOO_KEY}"}`, assert `inputs` entry generated with id `FOO_KEY`. | ~110 |
| 015 | Docs — update `specs/KNOWLEDGE_MAP.md` with spec 013 row, decisions (Copilot global probe-gated; skills→agents migration; MCP user-scope deep-merge; chatmodes VS-Code-only). Add Pending/Follow-up entries (Option C cloud path; `settings.local.json` equivalent if any). Append package additions to root `CLAUDE.md` codebase map. | ~40 |

**Exit criteria:** All new tests green; `go vet ./...` clean; KNOWLEDGE_MAP reflects the completed spec.

---

## 3. Scope hygiene

**In scope**
- All 11 defects enumerated in research §4.3 (G1–G11).
- Starter library assets: `library/copilot/agents/` + `library/copilot/instructions/`.
- Global scope support for Copilot via probe.
- MCP user-scope deep-merge; inputs scaffolding on env-placeholder detection.
- Skills → agents migration (abandons `.prompt.md` transform path).
- Post-install smoke test gated on CLI presence.

**Explicitly out of scope (per decisions §10)**
- Cloud / marketplace publishing (Q10 — parked follow-up).
- Chatmodes → agents migration (Q9 — chatmodes stay VS-Code-only, docs only).
- Non-language instruction files beyond the three shipped (typescript/go/tests).
- Other tools — Claude, OpenCode, Gemini, Codex — untouched.
- `copilot mcp add` interactive orchestration (doesn't exist non-interactively per research §3).

**Explicit non-goals**
- No ADR required. The global-scope flip is mechanical (mirrors patterns already in place for every other tool); nothing crosses a new architectural boundary. If Phase 4's deep-merge surfaces unexpected ambiguity we promote to ADR at that checkpoint.

---

## 4. Risk register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `.agent.yaml` schema drifts upstream before spec lands | Low | Medium | Phase 1 task 003 pins required fields in a schema test; failure surfaces immediately. |
| User has both VS Code `.github/agents/` and CLI `~/.copilot/agents/` — content divergence | Medium | Low | Global emitter copies from the same library source as project emitter; round-trip test asserts content equality between scopes (Phase 6). |
| Existing user-authored `~/.copilot/mcp-config.json` gets clobbered | Medium | High | Phase 4 task 010 uses `configmerge.MergeJSONFile` + `.bak` backup-on-first-touch (battle-tested pattern from OpenCode spec 011). |
| Skills migration (task 005) breaks users who relied on `skills/*.prompt.md` in VS Code | Medium | Medium | Commit message + KNOWLEDGE_MAP note + migration warning: on install, if `.github/prompts/<skill>.prompt.md` exists that was emitted by us, remove it and emit `.github/agents/<skill>.agent.yaml` in its place. Track via file-records owner. |
| `copilot` CLI changes its agent YAML schema | Low | Medium | Validation task 011 surfaces schema mismatches as warnings, not errors. User can still run VS Code. |
| Probe false negative — user has CLI installed to non-PATH location | Medium | Low | Fallback probe to `~/.copilot/` directory presence (task 007); covers the brew-installed-but-not-yet-run case. |
| Inputs scaffolding (task 006) emits duplicates when the same `${VAR}` appears in multiple servers | Medium | Low | De-duplicate by `id` before writing (unit test). |

---

## 5. Verification rounds (per CLAUDE.md protocol)

This is a **moderate-to-complex** task (15 tasks, 6 phases, new scope support) → **3 verification rounds** at the spec level:

1. Requirements re-check against `checklists/requirements.md` after each phase.
2. Edge-case pass at end of Phase 4: probe-fail path, re-install over existing `~/.copilot/`, workspace scope equivalence with project scope.
3. Integration-boundary pass at end of Phase 6: cross-tool scope parity test (is Copilot × global now equivalent to other tools × global in shape?), frontmatter schema conformance, MCP deep-merge preservation.

---

## 6. Open items that could reshape the plan

None blocking. The decisions in research §10 cover every branch point. Two monitors during implementation:

1. If `.agent.yaml` zod validation (via `copilot --agent <name> -p ping`) rejects a shape we emit in Phase 1, we stop and fix the library asset before continuing to Phase 2. Schema-test-first (task 003) should catch this before task 011 runs it live.
2. If `configmerge.MergeJSONFile` can't handle the `mcp-config.json` nesting (servers-as-map-of-objects), Phase 4 may need a small per-server-merge helper similar to OpenCode's. Mitigation: reuse the exact OpenCode MCP deep-merge pattern (`specs/011-opencode-deep-setup/tasks/004-mcp-per-server-deep-merge.md`).

---

## 7. Task files

See `tasks/001-*.md` through `tasks/015-*.md`. Each file is self-contained: goal, files to touch, acceptance criteria, test plan.
