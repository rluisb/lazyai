# Spec 025: LazyAI Runtime Refactor

## Evidence-Gated Runtime Debloat and Neutral Adapter Contract

**Status:** Draft — RPI research/specification gate pending. ADR-003 is proposed until a human-authored tracked approval commit verifies acceptance. P0 artifacts are currently untracked; no implementation is authorized.
**Author:** Ricardo Conceicao  
**Date:** 2026-06-13  
**Scope:** Destructive refactor of LazyAI runtime, adapters, persistence, handoff, and canonical library governance  
**Depends on:** `specs/007-go-runtime-foundation/`, `specs/008-go-runtime-session/`, `specs/009-go-runtime-taskqueue/`, `specs/010-go-runtime-workflow/`, `specs/adrs/003-lazyai-runtime-boundary.md`

---

## §G Goal

Refactor `lazyai` into a compact Go runtime for agentic CLI tooling while preserving data, preserving adapter value, removing Fortnite/orchestrator bloat safely, and enforcing a small canonical library through measurable token-rent gates.

---

## §C Constraints

1. **Human-gated destructive refactor** — No implementation (code changes) proceeds until Phase 0 prerequisites are approved and confidence reaches ≥85%. The plan (`plan.md`) is a Phase 0 prerequisite artifact (P0-8), not a post-approval artifact; it exists as a pre-approval draft and is revised as prerequisite evidence arrives.
2. **Runtime boundary** — `lazyai` owns runtime behavior; `vibe-lab` supplies principles and adapter expectations, not runtime implementation.
3. **Evidence before deletion** — No package, table, CLI command path, adapter branch, or library asset is removed until its dependency audit and rollback path are documented.
4. **Corrected order only** — Execution order is Audit → Adapter rewrite → Test rewrite → Excision → Schema migration → Handoff → Library curation.
5. **No schema deletion without reversal** — Every persistence change requires V2 design, backup, restore verification, and synthetic FK-saturated migration tests. Rollback is backup-restore (`lazyai restore-runtime-db`), not SQL down migration (no `lazyai migrate down` command exists).
6. **Neutral adapters first** — Adapter behavior MUST be defined by a canonical contract before Fortnite-specific assumptions are removed from tests.
7. **Orchestrator paradox resolved up front** — The non-Fortnite adapter MUST either install a lightweight agent definition or use a redesigned primary agent path; it MUST NOT require the heavy orchestrator package after excision.
8. **Token-rent is enforced, not advisory** — Canonical library size MUST be checked by CI/pre-commit tooling and MAY be exceeded only through documented `.lazyai` override policy.
9. **Each phase leaves the repo working** — Every phase checkpoint must pass its relevant static, behavioral, migration, and rollback checks before the next phase begins.
10. **ADR mandatory** — Runtime boundary, Fortnite excision, and adapter ownership decisions require an ADR before plan approval.

---

## §E Evidence Inputs

| Source | Finding consumed |
|---|---|
| `docs/reports/final-review-deepseek-v4-pro-2026-06-13.md` | Current confidence 63%; 8 evidence-gated prerequisites; corrected 30–40 day estimate; unresolved orchestrator paradox; token-rent and rollback gaps. |
| `docs/reports/final-review-gemini-3.1-pro-2026-06-13.md` | Current confidence ~63.5%; 7 blocking prerequisites; schema underestimation; missing fictional designs; adapter test contract gap; dead `FortniteMode` path. |
| `docs/reports/lazyai-runtime-adversarial-synthesis-2026-06-13.md` | Strategic decision that `lazyai` is runtime and `vibe-lab` is not; original plan rejected until expanded into a human-approved implementation plan. |
| `.specify/memory/constitution.md` | Articles II, III, IV, V, VI and Gates 1–5 govern the refactor. |

---

## §I Interfaces and Artifact Contracts

### §I.1 Dependency Audit Matrix

| Field | Requirement |
|---|---|
| Scope | Every `packages/cli/cmd/*.go` command file and every runtime/orchestrator/Fortnite import path named by the reviews. |
| Classification | `breakage`, `rewrite`, `keep`, `remove`, or `defer-with-owner`. |
| Evidence | Import chain, call path, user-facing command impact, and test coverage for the chosen disposition. |
| Gate use | Blocks plan approval if any command is unclassified. |

### §I.2 Neutral Adapter Contract

| Field | Requirement |
|---|---|
| Scope | OpenCode, Claude Code, Copilot — the three adapters currently registered in `packages/cli/internal/adapter/registry.go`. Gemini and Codex are NOT in scope; their adapter implementations do not exist in the current codebase. If reintroduced later, they follow the same neutral contract pattern established here.
| Contract | Expected files, agents, defaults, generated commands, config merge semantics, and validation behavior per adapter. |
| Exclusions | Fortnite-specific names, orchestration defaults, and hidden fallback assumptions unless explicitly justified. |
| Gate use | Adapter tests are rewritten against this contract before package excision. |

### §I.3 V2 Runtime Schema Contract

| Field | Requirement |
|---|---|
| Scope | Runtime state that remains after taskqueue/workflow/orchestrator removal. |
| Evidence | ER diagram, DDL, backup/restore path, synthetic dataset, FK cascade behavior, default-agent migration handling. |
| Gate use | Blocks schema implementation if V1 → V2 → V1 round-trip restore is not verified. |

### §I.4 Handoff Markdown Contract

| Field | Requirement |
|---|---|
| Scope | Session handoff files written by the runtime. |
| Evidence | Frontmatter keys, required sections, path rules, ownership model, append/update semantics, and round-trip tests. |
| Gate use | Blocks `WriteHandoff(path string)` implementation until approved. |

### §I.5 Token-Rent Budget Contract

| Field | Requirement |
|---|---|
| Scope | Canonical library agents, skills, hooks, commands, and tool-specific generated assets. |
| Evidence | Byte-count rule, default budget, CI/pre-commit integration, failure output, `.lazyai` documented exception format. |
| Gate use | Blocks library curation completion if overflow is not caught automatically. |

### §I.6 Rollback Contract

| Field | Requirement |
|---|---|
| Scope | Every destructive phase. |
| Evidence | Tagged release point, archived assets, down migration or restore path, user notification template, verification command, and owner. |
| Gate use | Blocks phase completion if rollback cannot be performed without reading source code. |


### §I.7 Primary-Agent Path Contract

| Field | Requirement |
|---|---|
| Scope | The non-Fortnite adapter path that replaces the orchestrator concept after `packages/orchestrator/` and `packages/cli/internal/orchestrator/` are removed. |
| Decision | **Option B: Redesigned primary-agent path** (per Clarify Q1 resolution). No lightweight `orchestrator.md` shim. Non-Fortnite adapters wire directly to a dispatcher or task-agent path defined in the canonical library. |
| Contract | The primary-agent path MUST be a single agent definition file in `packages/cli/library/canonical/agents/primary-agent.md` with: (a) a `name` field, (b) a `description` field, (c) a `tools` list, (d) no imports of removed orchestrator packages. |
| Migration | CLI commands that previously referenced `orchestrator` (e.g., `cmd/orchestration.go`, `cmd/task.go`, `cmd/workflow.go`) are rewritten to use the primary-agent path or removed with migration notes. |
| File outputs | The primary-agent path produces: (1) the canonical agent definition file, (2) updated adapter install logic in `opencode.go`/`claudecode.go`/`copilot.go` that references it instead of `orchestrator`, (3) updated `cmd/helpers.go` removing `FortniteMode` toggle. |
| Gate use | Blocks Phase 2 excision until the primary-agent definition exists and adapter tests pass against it. |
---

## §V Invariants

1. **No implementation before confidence floor** — Implementation MUST NOT start while assessed confidence is below 85%.
2. **No orphaned CLI command** — Every existing CLI command path MUST either keep working, be intentionally removed with user-facing migration notes, or be explicitly deferred with an owner AND a concrete deadline AND remain functional during the deferral period. A deferred command that is broken is a violation.
3. **No Fortnite assumption in neutral tests** — Adapter tests MUST assert canonical adapter behavior, not Fortnite library behavior.
4. **No heavy orchestrator dependency after excision** — Runtime and non-Fortnite adapter paths MUST NOT import the removed orchestrator packages.
5. **No data-loss migration** — V1 data MUST survive V2 migration or fail safely with rollback instructions and untouched backup.
6. **No advisory-only budget** — Library budget checks MUST fail automatically when the default budget is exceeded without a documented override.
7. **No hidden defaults** — Runtime defaults MUST NOT silently preserve `loop-driver`, `orchestrator`, or Fortnite defaults unless the approved contract names them.
8. **No archive-only rollback** — Archival alone MUST NOT satisfy rollback readiness.
9. **No phase skip** — Plan, tasks, analyze, checklist, and implementation phases remain blocked until their human gates are satisfied.

---

## §T Evidence-Gated Workstreams

| ID | Workstream | Evidence required | Gate |
|---|---|---|---|
| T001 | CLI import audit | Per-file matrix covering task, workflow, orchestration, MCP setup, config, helpers, doctor, add, message, validate, server, list, and update commands. | Phase 0 |
| T002 | Usage survey | Active Fortnite/OpenCode usage, workflow usage, migration blockers, and notification needs. | Phase 0 |
| T003 | V2 schema design | ER diagram, DDL, backup/restore path, default-agent mapping, synthetic FK-saturated test data. | Phase 0 |
| T004 | Handoff schema design | Frontmatter spec, sections, path conventions, ownership model, and round-trip criteria. | Phase 0 |
| T005 | Canonical library spec | Agent/skill/hook/command inventory with usage justification and byte budget contribution. | Phase 0 |
| T006 | Adapter contract rewrite | Neutral adapter behavior table plus rewritten tests for each adapter mode. | Phase 1 |
| T007 | Orchestrator paradox resolution | Human-approved decision: lightweight `orchestrator.md` agent definition or redesigned primary agent path. | Phase 1 |
| T008 | CLI excision and archive | Rewritten/removed commands, archived Fortnite library, removed orchestrator/runtime workflow packages, and stale-reference cleanup. | Phase 2 |
| T009 | Schema migration execution | V2 migration, backup restore verification, and session/ledger/handoff round-trip tests. | Phase 3 |
| T010 | Handoff implementation | Runtime handoff write/read behavior matching approved markdown schema. | Phase 4 |
| T011 | Token-rent enforcement | CI/pre-commit budget check and `.lazyai` exception policy. | Phase 5 |
| T012 | Rollback readiness | Tagged release, restore command, user notification template, and per-phase rollback verification. | Every phase |

---

## §P Corrected Phase Plan

### Phase 0 — Prerequisites and approval gate

**Estimate:** 5–7 working days  
**Exit:** All eight prerequisite artifacts exist, are linked, and are approved by a human.

Deliverables:
1. CLI command import audit.
2. Fortnite/OpenCode usage survey.
3. V2 runtime schema design.
4. Handoff markdown schema.
5. Canonical library specification.
6. Token-rent enforcement design.
7. Corrected dependency-order plan.
8. Rollback procedure.

### Phase 1 — Adapter decouple and test rewrite

**Estimate:** 10–14 working days  
**Exit:** Adapter behavior is neutral, tested, and no longer depends on Fortnite assumptions.

Deliverables:
1. Adapter test contract.
2. Rewritten adapter tests.
3. Orchestrator paradox decision applied.
4. Dead Fortnite mode/default paths removed or redesigned.

### Phase 2 — CLI command rewrite and excision

**Estimate:** 8–10 working days  
**Exit:** Affected CLI commands are rewritten or removed, heavy orchestration packages are gone, and user-facing command behavior is covered.

Deliverables:
1. Rewritten/removed task, workflow, and orchestration commands.
2. Removal of heavy orchestrator/runtime workflow/taskqueue/dispatch surfaces according to audit disposition.
3. Archive of Fortnite library content for rollback/reference.
4. Cleanup of stale references in config, helpers, doctor, add, message, validate, server, list, and update flows.

### Phase 3 — V2 schema migration

**Estimate:** 8–12 working days  
**Exit:** V1 runtime data migrates to V2 and can be restored or reversed under test.

Deliverables:
1. V2 migration up.
2. Backup/restore verification.
3. Migration reversal testing.
4. Session/ledger/handoff round-trip tests.

### Phase 4 — Handoff implementation

**Estimate:** 3–4 working days  
**Exit:** Runtime writes handoffs matching the approved markdown schema.

Deliverables:
1. Handoff write behavior.
2. Session integration.
3. Round-trip validation.
### Phase 5 — Library curation and enforcement

**Estimate:** 3–4 working days  
**Exit:** Canonical library stays within default budget or fails with documented override instructions.

Deliverables:
1. Curated canonical agent/skill/hook/command set.
2. CI/pre-commit byte-budget check.
3. `.lazyai` documented exception support.


**Total realistic estimate:** 37–51 working days (sum of phase estimates: 5–7 + 10–14 + 8–10 + 8–12 + 3–4 + 3–4).

---

## §B Backward Compatibility and Rollback

1. Existing user data MUST be backed up before migration.
2. Migration tests MUST verify V1 → V2 and V2 → V1 or restore-from-backup behavior.
3. Removed library/runtime assets MUST be archived before deletion but archive alone does not satisfy rollback.
4. Each phase MUST name the rollback trigger, command, owner, and verification evidence.
5. User-facing command removals MUST include migration notes or explicit deprecation/removal messaging.
6. Tagged releases MUST mark pre-refactor restore points before destructive phases.

---

## §A Acceptance Criteria

1. Phase 0 approval is impossible until all eight prerequisite artifacts are present and linked from the plan.
2. The CLI audit classifies every `packages/cli/cmd/*.go` file and every reported orchestrator/runtime/Fortnite dependency.
3. The orchestrator paradox has exactly one human-approved resolution, and adapter tests encode that resolution.
4. Adapter tests pass against the neutral contract without requiring Fortnite library content.
5. V2 schema migration passes on synthetic data covering FK cascades, partial unique indexes, legacy defaults, and empty databases.
6. Backup/restore or down migration evidence exists before any schema-shrinking change lands.
7. Handoff behavior conforms to the approved markdown schema and round-trips through session workflows.
8. Token-rent tooling fails the build when the canonical library exceeds the default budget without a documented override.
9. Every destructive phase has a rollback verification record and user notification template.
10. Gate 5 evidence proves operators can detect and roll back the refactor without reading source.

---

## §Q Clarifications (Resolved 2026-06-13)

1. **Orchestrator paradox → B: Redesigned primary-agent path.** Remove the orchestrator concept entirely. Non-Fortnite adapters wire to a redesigned primary agent (dispatcher or direct task-agent path). No `orchestrator.md` shim.
2. **Usage survey source → GitHub issues + direct user survey.** No CLI telemetry exists. Authoritative evidence: (a) GitHub issues tagged `fortnite`/`opencode` for quantitative baseline, (b) direct outreach (Slack/Discourse/github discussions) for qualitative blockers and migration sentiment.
3. **Default token budget → Option A: 50KB (50,000 bytes) project-wide.** Single budget across all canonical library categories (agents + skills + hooks + commands). Escape hatch: `.lazyai/token-rent-override` with required `reason:` field. CI/pre-commit enforces via `wc -c` pipeline.
---

## §N Next Steps

1. Human reviews and approves this spec and ADR-003.
2. After spec/ADR approval, create all eight Phase 0 prerequisite artifacts with concrete evidence (audit, survey, schema designs, contracts).
3. Revise `plan.md` (P0-8) to incorporate evidence from completed P0-1 through P0-7 artifacts.
4. ⛔ Human approves the revised plan (Phase 0 gate).
5. After plan approval, generate `tasks.md`, `analysis.md`, and `checklists/`.
6. After tasks/analyze/checklists pass, begin implementation one phase at a time with human gates.

---

## §R References

- DeepSeek final review: `docs/reports/final-review-deepseek-v4-pro-2026-06-13.md`
- Gemini final review: `docs/reports/final-review-gemini-3.1-pro-2026-06-13.md`
- Synthesis artifact: `docs/reports/lazyai-runtime-adversarial-synthesis-2026-06-13.md`
- Constitution: `.specify/memory/constitution.md`
