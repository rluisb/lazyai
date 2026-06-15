# LazyAI → Vibe-Lab Runtime: Adversarial Review & Refactor Report

**Date:** 2026-06-13
**Scope:** Review LazyAI's agents, workflows, skills, hooks, and OpenCode-heavy runtime through vibe-lab's Akita philosophy lens. Identify what to strip, what to refactor, what to reuse.
**Mode:** RESEARCH, REVIEW, ADVERSARIAL REVIEW, REPORT_ONLY — no implementation.

---

## §1 Vibe-Lab Philosophy Baseline

From `concepts/prompt-engineering.md`, `concepts/context-engineering.md`, `concepts/harness-engineering.md`, `concepts/llm-engineering-relationships.md`, and the Akita adversarial reviews (2026-06-05):

### Core Principles

| Principle | Source | Meaning |
|-----------|--------|---------|
| **Four-Point Contract** | Akita | WHAT, HOW, DON'T WANT, VALIDATE — minimal, no frameworks |
| **No heavy frameworks** | Akita | No YAML workflow engines, dispatch layers, coordinator agents |
| **Add only when needed** | Akita | Defer until concrete use case exists |
| **Explicit non-goals** | Akita | Prevents scope creep |
| **One task at a time** | Akita | Single-agent sessions need no orchestration overhead |
| **Minimal markdown/scripts/symlinks** | Akita | Over frameworks |
| **No speculative abstractions** | Akita | Build what's needed, not what might be |
| **Clean code for agents** | Akita | Test-first, headless tests |
| **Concrete validation** | Akita | Don't claim enforcement without proof |

### Engineering Boundaries

| Layer | Owns | Anti-Pattern |
|-------|------|-------------|
| **Prompt Engineering** | How to ask — role, task, constraints, output shape | Vague asks, mixing instructions with data, elaborate scaffolds for simple tasks |
| **Context Engineering** | What model should know right now — selection, compression, injection | Loading everything, stale docs, treating context as infinite |
| **Harness Engineering** | How work executes/verifies — contracts, sensors, state, gates | No handoff, skipping verification, letting work continue after failures |

### The Bridge Pattern

The four-point framework **scales into** orchestration when needed. The **pattern** is portable; the **framework** is deferred. Don't adopt OpenCode's framework — adopt its patterns as lightweight hooks and skills.

---

## §2 LazyAI Current State — Inventory

### 2.1 OpenCode-Heavy Runtime (Fortnite)

`packages/cli/library/fortnite/` is the **production OpenCode runtime** embedded in LazyAI:

| Component | Count | Details |
|-----------|-------|---------|
| Fortnite Agents | 8 | loop-driver, engine-control, loot-hawk, turbo-crank, wall-builder, shield-audit, rift-deploy, respawn-crew |
| Fortnite Skills | 40+ | battle-bus, build-mode, zero-point, truth-chain, war-council, workflow-engine, storm-scout, etc. |
| Fortnite Workflows | 8 | bugfix, hotfix, refactor, review-mine, review-others, rpi, spike, test-failure-policy |
| Bash Scripts | 15+ | session-db.sh (79KB), task-queue.sh, workflow-run.sh, run-evals.sh, etc. |
| Support Docs | 7 | DISPATCH-MATRIX, FALLBACK-CHAINS, SAFETY-BOUNDARIES, TOOL-SCHEMAS, etc. |

### 2.2 Go Runtime Migration (In Progress)

Specs 007–010 are porting the bash runtime to Go:

| Spec | What | Status |
|------|------|--------|
| 007 | Foundation — unified schema, DB layer, migrations | Draft |
| 008 | Session — lifecycle, dispatches, parallel tasks, barriers, locks, messaging | Draft |
| 009 | Task Queue — atomic claiming, DLQ, zombie sweep | Draft |
| 010 | Workflow Engine — YAML parsing, phase dispatch, human gates, agent dispatch | Draft |

### 2.3 Adapter Layer

`packages/cli/internal/adapter/` compiles canonical `.ai/` sources to per-CLI formats:

| Adapter | File | Lines |
|---------|------|-------|
| OpenCode | `opencode.go` | ~420 |
| Claude Code | `claudecode.go` | ~530 |
| Copilot | `copilot.go` | ~460 |
| Shared | `shared.go` | ~960 |
| MCP Compiler | `mcp_compiler.go` | ~620 |

### 2.4 Agent/Skill/Hook Surface

| Location | Agents | Skills | Hooks |
|----------|--------|--------|-------|
| `.opencode/` | 8 | 40+ | 0 |
| `.claude/` | 32 (YAML) | 35+ | 2 |
| `.github/` | 7 | — | — |
| `.agents/` | — | 30+ | — |
| `packages/cli/library/agents/` | 8 (canonical) | — | — |
| `packages/cli/library/skills/` | — | 30+ (canonical) | — |
| `packages/cli/library/fortnite/agents/` | 8 (Fortnite) | — | — |
| `packages/cli/library/fortnite/skills/` | — | 40+ (Fortnite) | — |

**Total unique agents:** ~55 across all locations (many duplicates across CLI adapters)
**Total unique skills:** ~45 across all locations (many duplicates)
**Total hooks:** 2 (both Claude Code only)

### 2.5 Orchestrator Package

`packages/orchestrator/` — separate Go binary with ports/domain/adapters (hexagonal):
- Catalog store, workflow state, job queue, handoff store, budget tracker, agent invoker
- SQLite adapters, MCP integration, dispatch, events

### 2.6 Existing Gap Analysis

`GAP-ANALYSIS.md` (2026-05-23) already identified 12 capability gaps:
- 5 missing: agent contracts, session tracking, audit trail, eval harness, model routing
- 4 partial: skills, health checks, smoke tests, safety boundaries
- 3 present: workflow engine, speckit integration, infrastructure

---

## §3 Adversarial Review — Vibe-Lab Lens

### 3.1 CRITICAL: OpenCode Is the Wrong Foundation

**Finding:** LazyAI's entire runtime identity is OpenCode/Fortnite. The Go migration (specs 007–010) is literally porting Fortnite bash scripts to Go. The adapter layer compiles canonical sources into OpenCode/ClaudeCode/Copilot formats. The orchestrator package is a separate hexagonal Go binary.

**Vibe-Lab Verdict:** This is the **opposite** of vibe-lab's philosophy.

| Vibe-Lab Principle | LazyAI Reality | Gap |
|-------------------|---------------|-----|
| No heavy frameworks | Fortnite = 8 agents, 40+ skills, 8 workflows, dispatch matrix, fallback chains, safety boundaries, tool schemas, repo profiles, output schemas | **MASSIVE** |
| Add only when needed | Everything is pre-built: workflow engine, task queue, session DB, eval harness, model routing, orchestrator | **MASSIVE** |
| One task at a time | Multi-agent dispatch with parallel execution, barriers, locks | **MASSIVE** |
| Minimal markdown/scripts | 79KB session-db.sh, 20KB workflow-run.sh, 15+ bash scripts | **MASSIVE** |
| No speculative abstractions | Hexagonal orchestrator with ports/domain/adapters for a personal lab | **HIGH** |
| Concrete validation | GAP-ANALYSIS shows 5 capabilities missing, 4 partial | **HIGH** |

**Recommendation:** Strip the Fortnite runtime entirely from LazyAI. The Go runtime migration (specs 007–010) should be **halted** — it's porting the wrong foundation. The orchestrator package should be **archived** — hexagonal architecture for a personal lab is speculative over-engineering.

### 3.2 Agent Surface Is Bloated and Duplicated

**Finding:** ~55 agents across 5 locations with significant duplication. `.claude/` has 32 YAML agents; `.opencode/` has 8 markdown agents; `.github/` has 7; the canonical library has 8; Fortnite has 8. Many are the same agent in different CLI formats.

**Vibe-Lab Verdict:** This violates context engineering ("load only task-relevant context") and prompt engineering ("avoid conflicting prompt layers"). An agent shouldn't need to know about 55 possible sub-agents.

**Vibe-Lab's approach:** 7 agents in vibe-lab, each with a clear `Use when...` trigger. The canonical → adapter pattern is correct, but the surface is too large.

**Recommendation:** Reduce to ≤8 core agents. Keep only: researcher, planner, implementor, reviewer, scout (code exploration), and 2–3 domain-specific agents. Drop: orchestrator, parallel-execution, process-audit, anti-speculation, self-improve, impact-check, extract-standards, housekeeping, memory-write, update-memory, chain-verify, catalog-manage, proof-of-concept, red-team-plan, diagnose, investigate, iterate, improve-codebase-architecture. These are either orchestration agents (anti-Akita) or meta-agents that add prompt noise.

### 3.3 Skill Surface Is a Token Bomb

**Finding:** ~45 skills across locations. Fortnite alone has 40+ skills including battle-bus (32KB), graphify (56KB), drop-ship (18KB), build-fort (15KB), zero-point (14KB), storm-scout (14KB).

**Vibe-Lab Verdict:** This is the **context over-injection** problem identified in every vibe-lab audit. Generated AGENTS.md/CLAUDE.md inline all skill catalogs. 45 skills in default context = massive prompt noise.

**Vibe-Lab's approach:** 21 skills, but the audits flagged even that as too many without a token-rent gate. The fix recommended: split runtime core from on-demand references.

**Recommendation:** Reduce to ≤15 skills. Keep only the essential workflow skills (speckit-specify, speckit-plan, speckit-tasks, speckit-implement, speckit-checklist, speckit-analyze, speckit-clarify) plus core utilities (bugfix, review, spike, research). Drop all orchestration/meta skills (battle-bus, workflow-engine, task-queue, parallel-execution, orchestrate, process-audit, war-council, truth-chain, zero-point, storm-scout, storm-plan, storm-research, storm-clarify, storm-eye, drop-ship, ricochet, lesson-loot, slurp-juice, the-vault, supply-llama, build-fort, shield-wall, build-mode, compact-blueprint, feedback-review, pr-review, pr-watcher, meeting-prep, daily-standup, dev-cli, dev-health, colima, hotctl, refresh-dev-containers, reboot-van, drift-scope, sidecar, worktree-manager, qmd, graphify, inventory-shrink, implement, iterate, catalog-manage, chain-verify, dynamic-compose, capture-tasks-from-meeting-notes, generate-status-report, spec-to-backlog, search-company-knowledge, triage-issue, a11y-debugging, debug-optimize-lcp, memory-leak-debugging, chrome-devtools, chrome-devtools-cli, troubleshooting, teachable-dev-playbook, teachable-school-automation, lazyai-design-system, tui-lazy-ai-design-system).

### 3.4 Hooks Are Nearly Absent

**Finding:** Only 2 hooks, both Claude Code only (`rpi-gate-check.yml`, `pre-commit`). No OpenCode hooks, no Copilot hooks, no safety boundaries beyond pre-commit.

**Vibe-Lab Verdict:** This is the **safety boundary gap** identified in the harness audit. Vibe-lab's audits found: single hook with string matching is trivially evaded; POLICY.md is advisory (fail-open), not enforced.

**Recommendation:** Add a PreToolUse allow-list hook (default-deny with explicit allow-list) — the Phase 1 recommendation from the vibe-lab harness audit. ~60 lines. This is the single highest-ROI safety addition.

### 3.5 Workflow Engine Is Premature

**Finding:** LazyAI has a full YAML workflow engine (spec 010) with phase dispatch, human gates, variable interpolation, fallback handling, and agent dispatch. Plus 8 Fortnite workflows. Plus the orchestrator package with catalog/workflow state management.

**Vibe-Lab Verdict:** Workflow orchestration is **correctly excluded** from vibe-lab per the harness audit. The Akita principle: "No YAML workflow engine, no dispatch layer, no coordinator agent." The bridge pattern says the four-point framework scales into orchestration when needed — but the framework is deferred.

**Recommendation:** **Remove the workflow engine.** It's Phase 7 (last, lowest priority) in vibe-lab's adoption path. A personal lab doesn't need YAML workflow execution. The four-point contract (WHAT, HOW, DON'T WANT, VALIDATE) is the workflow.

### 3.6 Task Queue Is Premature

**Finding:** Spec 009 implements SQLite-backed task queue with atomic claiming (CTE + RETURNING), dead-letter queue, zombie sweeper, and task-scoped chat.

**Vibe-Lab Verdict:** Multi-agent task queues are orchestration infrastructure. Vibe-lab explicitly defers multi-agent workflows. A single-agent personal lab doesn't need atomic task claiming.

**Recommendation:** **Remove the task queue.** If multi-agent work becomes a concrete need later, add it then — not before.

### 3.7 Session Tracking Is Useful But Over-Engineered

**Finding:** Spec 008 implements session lifecycle, dispatch tracking, parallel task tracking, messaging, barriers, and locks — all ported from the 79KB `session-db.sh`.

**Vibe-Lab Verdict:** Session tracking is useful (handoff protocol, audit trail). But barriers, locks, parallel task tracking, and dispatch tracking are orchestration infrastructure for multi-agent systems.

**Recommendation:** **Keep session CRUD only.** Drop dispatch tracking, parallel tasks, barriers, locks, and messaging. A single-agent session needs: start, log prompt/response, end. That's it.

### 3.8 Adapter Layer Is the Right Pattern — But Wrong Surface

**Finding:** The canonical-source → compile model (`packages/cli/internal/adapter/`) is **strongly Akita-aligned**. This is exactly vibe-lab's `bin/inject` + `bin/doctor` pattern. The problem is WHAT it compiles: 55 agents, 45 skills, 8 workflows, Fortnite runtime.

**Vibe-Lab Verdict:** Keep the adapter pattern. It's the innovation. But compile a vibe-lab-sized surface (≤8 agents, ≤15 skills, 0 workflows, 2–3 hooks).

**Recommendation:** **Preserve the adapter layer.** It's the one piece of LazyAI that maps directly to vibe-lab's canonical → adapter architecture. Strip what it compiles, not the compiler.

### 3.9 The Orchestrator Package Is Speculative Over-Engineering

**Finding:** `packages/orchestrator/` is a separate Go binary with hexagonal architecture (ports/domain/adapters), SQLite adapters, MCP integration, catalog management, budget tracking, job queues, handoff stores.

**Vibe-Lab Verdict:** This is the definition of speculative abstraction. A personal lab doesn't need hexagonal architecture, budget trackers, or catalog stores. The Akita principle: "No speculative abstractions."

**Recommendation:** **Archive the orchestrator package.** It may have value for a future multi-tenant SaaS product, but it has no place in a personal AI lab runtime.

### 3.10 Documentation Drift

**Finding:** README claims LazyAI is a "CLI-first AI operating system for software teams" with session management, health checks, audit trail, validation, workspace management, task queue, agent message bus, git integration, backup, secrets, notifications, metrics dashboard, completion, memory vault, evaluation harness, workflow execution, and sidecar. The GAP-ANALYSIS shows 5 of 12 capabilities are missing.

**Vibe-Lab Verdict:** This is the **docs-drift-from-reality** problem identified in every vibe-lab audit. Claims exceed implementation.

**Recommendation:** Rewrite README to match the stripped-down reality. One paragraph: "LazyAI is a personal AI engineering lab. It compiles a small set of canonical agents, skills, and hooks into your chosen CLI tool. No orchestration, no workflow engines, no task queues — just the four-point contract and concrete validation."

---

## §4 What to Keep (Vibe-Lab Aligned)

| Component | Verdict | Rationale |
|-----------|---------|-----------|
| **Adapter layer** (`internal/adapter/`) | **KEEP** | Canonical → compile is the core innovation. Maps directly to vibe-lab's `bin/inject`. |
| **Canonical library** (`library/agents/`, `library/skills/`) | **KEEP, SHRINK** | Right pattern, wrong size. Reduce to ≤8 agents, ≤15 skills. |
| **Session CRUD** | **KEEP, SIMPLIFY** | Start, log, end. Drop dispatch/parallel/barriers/locks/messaging. |
| **Ledger** (`internal/runtime/ledger/`) | **KEEP** | Immutable append-only audit trail. Maps to vibe-lab's handoff protocol. |
| **DB layer** (`internal/runtime/db.go`, `schema.go`) | **KEEP, SHRINK** | Keep sessions + ledger tables only. Drop task_queue, workflows, teams, dispatches, parallel_tasks, barriers, locks, messages, model_calls. |
| **Sidecar** | **KEEP** | Optional, lightweight, local-only. Akita-aligned: add only when needed. |
| **MCP compiler** | **KEEP** | Compiling MCP config from canonical source is useful. |
| **Health checks** (`lazyai-cli doctor`) | **KEEP** | Maps to vibe-lab's `bin/doctor`. Must actually verify claims. |

---

## §5 What to Strip (Anti-Vibe-Lab)

| Component | Verdict | Rationale |
|-----------|---------|-----------|
| **Fortnite runtime** (`library/fortnite/`) | **REMOVE** | OpenCode-specific multi-agent system. Anti-Akita: heavy framework, pre-built orchestration. |
| **Workflow engine** (spec 010, `internal/runtime/workflow/`) | **REMOVE** | YAML workflow execution is Phase 7 in vibe-lab adoption path. Defer until needed. |
| **Task queue** (spec 009, `internal/runtime/taskqueue/`) | **REMOVE** | Multi-agent atomic claiming is orchestration infrastructure. |
| **Orchestrator package** (`packages/orchestrator/`) | **ARCHIVE** | Hexagonal architecture for personal lab = speculative over-engineering. |
| **Dispatch layer** (`internal/runtime/dispatch/`) | **REMOVE** | Multi-agent dispatch with parallel execution. |
| **Parallel execution** (session/parallel.go) | **REMOVE** | Multi-agent infrastructure. |
| **Barriers & locks** (session/barrier.go, session/lock.go) | **REMOVE** | Multi-agent coordination. |
| **Agent message bus** (session/message.go) | **REMOVE** | Inter-agent messaging. |
| **Model routing** (MODEL-ROUTING.md) | **REMOVE** | Defer until multi-model use case exists. |
| **Eval harness** (`.specify/evals/`) | **REMOVE** | Storm-eye eval framework is orchestration infrastructure. |
| **Bash scripts** (`library/fortnite/scripts/`) | **REMOVE** | 79KB session-db.sh, 20KB workflow-run.sh — ported to Go or dropped. |
| **Fortnite agents** (8 agents) | **REMOVE** | loop-driver, engine-control, loot-hawk, turbo-crank, wall-builder, shield-audit, rift-deploy, respawn-crew. |
| **Fortnite skills** (40+ skills) | **REMOVE** | battle-bus, build-mode, zero-point, truth-chain, war-council, etc. |
| **Fortnite workflows** (8 workflows) | **REMOVE** | bugfix, hotfix, refactor, review-mine, review-others, rpi, spike, test-failure-policy. |
| **Fortnite support docs** | **REMOVE** | DISPATCH-MATRIX, FALLBACK-CHAINS, SAFETY-BOUNDARIES, TOOL-SCHEMAS, OUTPUT-SCHEMAS, REPO-PROFILES. |
| **Excess agents** (~47 of 55) | **REMOVE** | Keep ≤8. Drop all orchestration/meta agents. |
| **Excess skills** (~30 of 45) | **REMOVE** | Keep ≤15. Drop all orchestration/meta/domain-specific skills. |
| **`.opencode/workflows/`** | **REMOVE** | bugfix.yaml, hotfix.yaml, rpi.yaml. |
| **`.agents/workflows/`** | **REMOVE** | Same workflows, different location. |
| **`.specify/templates/`** | **SHRINK** | Keep spec, plan, tasks templates. Drop audit, housekeeping, ledger, poc, checklist, spike, task-harness. |
| **`.github/agents/`** | **REMOVE** | Copilot-specific agents — defer until Copilot is a concrete target. |
| **`.github/chatmodes/`** | **REMOVE** | Copilot-specific. |
| **`.github/prompts/`** | **REMOVE** | Copilot-specific. |
| **`.github/instructions/`** | **REMOVE** | Copilot-specific. |
| **`demo/`** | **REMOVE** | 13 VHS tapes — demo artifacts, not runtime. |
| **`mkdocs.yml`** | **REMOVE** | Documentation site config — defer. |
| **`dashboard.html`**, **`metrics.prom`** | **REMOVE** | Monitoring infrastructure for multi-agent systems. |
| **`cupcake.yml`** | **REMOVE** | Policy engine — defer until safety boundaries are concrete. |
| **`policies/`** | **REMOVE** | Claude/OpenCode/common policies — defer. |

---

## §6 Refactor Plan — What to Reuse

### 6.1 Adapter Layer → `bin/inject` Equivalent

The adapter layer (`internal/adapter/`) is the **canonical-source → compile** engine. This is vibe-lab's `bin/inject` pattern, but in Go instead of bash.

**Reuse:** Keep `opencode.go`, `claudecode.go`, `shared.go`, `mcp_compiler.go`. Strip the Fortnite-specific install paths. The adapter should compile from `library/agents/` and `library/skills/` (canonical sources) into `.opencode/`, `.claude/`, etc.

**Refactor needed:**
- Remove `InstallToolContextFiles` (installs Fortnite agents-dir.md, skills-dir.md, root-dir.md)
- Remove orchestrator agent special-casing (`IsOrchestratorEnabled`, `ReadOrchestratorTools`, etc.)
- Remove Copilot adapter (defer until Copilot is a concrete target)
- Simplify `Install()` to: copy agents, copy skills, write managed block to root AGENTS.md, merge settings

### 6.2 Session DB → Handoff Protocol

The session CRUD (`internal/runtime/session/session.go`) can be simplified to a handoff protocol.

**Reuse:** `Start()`, `End()`, `Get()` — basic session lifecycle. The `session.db` SQLite file.

**Refactor needed:**
- Drop `Dispatch()`, `CreateParallelTask()`, `SendMessage()`, `CreateBarrier()`, `AcquireLock()`
- Add `WriteHandoff()` — writes goal, status, files changed, commands run, validation results, blockers, next step to `.vibe-lab/handoffs/`
- Schema: keep `sessions` table only. Drop `dispatches`, `parallel_tasks`, `messages`, `barriers`, `locks`, `task_queue`, `task_messages`, `workflows`, `workflow_steps`, `teams`, `model_calls`

### 6.3 Ledger → Truth Chain

The ledger (`internal/runtime/ledger/ledger.go`) is an immutable append-only JSONL ledger with hash-chain verification.

**Reuse:** Keep as-is. This maps to vibe-lab's handoff protocol recommendation: "truth-chain → 10-line SHA reference in a handoff file."

**Refactor needed:** None. The ledger is already the right shape.

### 6.4 Agent/Skill Library → Shrink to Vibe-Lab Core

**Reuse:** The canonical markdown format for agents and skills. The YAML frontmatter pattern.

**Refactor needed:**
- Keep ≤8 agents: researcher, planner, implementor, reviewer, scout, builder, documenter, red-team
- Keep ≤15 skills: speckit-specify, speckit-plan, speckit-tasks, speckit-implement, speckit-checklist, speckit-analyze, speckit-clarify, speckit-constitution, bugfix, review, spike, research, tdd-loop, rpi, update-memory
- Drop all others
- Add instruction/data boundary block to every agent/skill template (vibe-lab audit finding)
- Add explicit `Use when...` triggers to avoid broad auto-activation

### 6.5 MCP Compiler → Keep

**Reuse:** `mcp_compiler.go` compiles `.ai/mcp.json` into per-CLI MCP config. This is useful and lightweight.

**Refactor needed:** Remove orchestrator MCP server special-casing. Remove Copilot MCP compilation.

### 6.6 Sidecar → Keep

**Reuse:** The sidecar mechanism (spec 001) is optional, local-only, lightweight. Akita-aligned.

**Refactor needed:** None. Already the right shape.

---

## §7 Priority Matrix

| Rank | Action | Effort | Impact | Akita Alignment |
|------|--------|--------|--------|-----------------|
| **P0** | Strip Fortnite runtime entirely | High (remove ~50 files) | **Critical** | Removes the anti-pattern foundation |
| **P0** | Halt Go runtime migration (specs 007–010) | Zero (stop work) | **Critical** | Stops porting the wrong foundation |
| **P0** | Archive orchestrator package | Low (move dir) | **High** | Removes speculative over-engineering |
| **P1** | Shrink agent surface to ≤8 | Medium | **High** | Context engineering: load only what's needed |
| **P1** | Shrink skill surface to ≤15 | Medium | **High** | Context engineering: stop over-injection |
| **P1** | Add PreToolUse safety hook | Low (~60 lines) | **Critical** | Safety boundary hardening |
| **P1** | Simplify session to CRUD + handoff | Medium | **High** | Drops orchestration infrastructure |
| **P2** | Simplify DB schema (sessions + ledger only) | Low | **Medium** | Removes unused tables |
| **P2** | Add instruction/data boundary to templates | Low (~5 lines each) | **Medium** | Prompt engineering hygiene |
| **P2** | Rewrite README to match reality | Low | **Medium** | Docs truth maintenance |
| **P3** | Remove Copilot adapter | Low | **Low** | Defer until concrete target |
| **P3** | Remove model routing | Low | **Low** | Defer until multi-model use case |
| **P3** | Remove eval harness | Low | **Low** | Defer until quality measurement need |

---

## §8 What the Result Looks Like

After refactor, LazyAI becomes:

```
lazyai/
  packages/cli/
    cmd/              # CLI commands: init, compile, doctor, session, sidecar
    internal/
      adapter/        # opencode.go, claudecode.go, shared.go, mcp_compiler.go
      runtime/
        db.go         # SQLite connection (WAL, FK, busy timeout)
        schema.go     # sessions + ledger tables only
        migrate.go    # migration framework
        session/      # session.go (start, get, end, handoff)
        ledger/       # ledger.go (append, verify, chain)
      types/          # shared types
      theme/          # TUI theme
      scaffold/       # init scaffolding
    library/
      agents/         # ≤8 canonical agents (markdown)
      skills/         # ≤15 canonical skills (markdown)
      hooks/          # 2–3 hooks (safety, pre-commit)
      embed.go        # embed.FS for library files
    AGENTS.md         # runtime rules
  .opencode/          # generated by adapter (not committed)
  .claude/            # generated by adapter (not committed)
  specs/              # feature specs, ADRs, standards
  docs/               # concepts, getting-started
  README.md           # one paragraph: personal AI engineering lab
```

**Gone:**
- `packages/cli/library/fortnite/` (entire directory)
- `packages/orchestrator/` (entire package)
- `packages/cli/internal/runtime/workflow/`
- `packages/cli/internal/runtime/taskqueue/`
- `packages/cli/internal/runtime/dispatch/`
- `packages/cli/internal/runtime/session/` (dispatch.go, parallel.go, barrier.go, lock.go, message.go)
- `packages/cli/internal/adapter/copilot.go`
- `packages/cli/internal/adapter/copilot_cli.go`
- `.opencode/workflows/`
- `.agents/workflows/`
- `.github/agents/`, `.github/chatmodes/`, `.github/prompts/`, `.github/instructions/`
- `demo/`
- `mkdocs.yml`, `dashboard.html`, `metrics.prom`, `cupcake.yml`, `policies/`
- ~47 agents, ~30 skills, all Fortnite scripts/docs

---

## §9 Risks & Watchouts

1. **Breaking existing LazyAI users.** The Fortnite runtime is what `lazyai-cli init --preset fortnite` installs. Removing it breaks that preset. Mitigation: version the change as LazyAI v2 with a migration guide.

2. **Losing the production runtime.** Fortnite is a working, battle-tested multi-agent system. Removing it means LazyAI no longer ships a production runtime — it ships a personal lab scaffold. This is the **intended direction** (vibe-lab philosophy), but it's a product identity change.

3. **Go runtime migration sunk cost.** Specs 007–010 represent design work already done. Halting them wastes that effort. Mitigation: the session CRUD and ledger parts are reusable; only the orchestration parts (task queue, workflow engine, dispatch) are wasted.

4. **Adapter layer depends on library surface.** Shrinking agents/skills from 55/45 to 8/15 means the adapter's `Install()` function compiles far fewer files. This is a simplification, not a breakage.

5. **Copilot adapter removal.** If Copilot is a concrete target for users, removing the adapter is premature. Mitigation: keep the code in an `archive/` directory for future reference.

---

## §10 Verification Checklist

Before any implementation begins:

- [ ] Confirm vibe-lab philosophy alignment with stakeholders
- [ ] Confirm Fortnite runtime removal is acceptable (product identity change)
- [ ] Confirm orchestrator package archival
- [ ] Confirm agent/skill shrink targets (≤8 agents, ≤15 skills)
- [ ] Confirm Go runtime migration halt
- [ ] Confirm adapter layer preservation
- [ ] Confirm Copilot adapter deferral

---

*Adversarial review complete. No implementation performed. Report only.*

---

# APPENDIX A: Independent Adversarial Challenge (2026-06-13)

**Reviewer:** Independent red-team pass
**Method:** Fresh code analysis, no reliance on prior report conclusions

---

## A.1 Evidence Summary — Ground Truth

| Metric | Value | Source |
|--------|-------|--------|
| **CLI package total Go** | 64,268 lines | `find packages/cli -name "*.go"` |
| **Adapter layer** | 10,059 lines (3,010 core, 7,049 tests) | `packages/cli/internal/adapter/` |
| **Runtime total** | 4,295 lines | `packages/cli/internal/runtime/` |
| **Runtime core (session CRUD + ledger + DB + schema)** | 1,331 lines | session.go + ledger.go + db.go + schema.go |
| **Runtime orchestration (workflow + taskqueue + dispatch + session extras)** | 2,964 lines | workflow/, taskqueue/, dispatch/, session/dispatch.go, parallel.go, barrier.go, lock.go, message.go |
| **Internal orchestrator** | 1,261 lines | `packages/cli/internal/orchestrator/` |
| **Orchestrator package (separate binary)** | 18,175 lines (99 Go files) | `packages/orchestrator/` |
| **Fortnite library** | 146 files, 28,608 lines (6,172 md + 7,184 sh + yaml) | `packages/cli/library/fortnite/` |
| **Canonical agents** | 8 files | `packages/cli/library/agents/` |
| **Canonical skills** | 33 files | `packages/cli/library/skills/` |
| **Fortnite agents** | 8 agents | `packages/cli/library/fortnite/agents/` |
| **Fortnite skills** | 37 directories | `packages/cli/library/fortnite/skills/` |
| **Fortnite bash scripts** | 27 files, 7,184 lines | `packages/cli/library/fortnite/scripts/` |
| **Fortnite Go references** | 41 occurrences in `internal/` | `grep -r "fortnite" internal/` |
| **loop-driver references** | 17 occurrences | `grep -r "loop-driver"` |
| **CLI orchestration commands** | 1,836 lines | orchestration.go + workflow.go + task.go + session.go + ledger.go + message.go |
| **Total removable code** | ~49,880 lines | Fortnite + orchestrator pkg + runtime orchestration + CLI cmds |
| **Code remaining after strip** | ~14,388 lines | Adapter + runtime core + scaffold + CLI commands |

---

## A.2 Challenges to Prior Report

### A.2.1 CHALLENGE: "Halt Go Runtime Migration" — Partially Wrong

**Prior claim:** "The Go runtime migration (specs 007–010) should be halted — it's porting the wrong foundation."

**Challenge:** Only **partially correct**. Specs 007 (foundation) and 008 (session) contain the **1,331 lines of useful runtime core** (session CRUD, ledger, DB layer, schema management). The foundation spec's schema has 20+ tables but only 4–5 are needed (sessions, schema_migrations, ledgers, decisions, artifacts). The work to extract and simplify is less than starting from scratch.

**Evidence:** `packages/cli/internal/runtime/session/session.go` shows clean `Start()`, `End()`, `Get()` methods — 269 lines, testable, already used by 16+ CLI commands. This is not wasted work.

**Revised recommendation:** Don't halt — **redirect**. Keep spec 007 foundation (DB layer, migrations, connection management) and spec 008 session CRUD only. Discard spec 009 (task queue) and spec 010 (workflow engine) entirely.

### A.2.2 CHALLENGE: "Keep ≤15 Skills" — Arbitrary Number

**Prior claim:** "Reduce to ≤15 skills. Keep only the essential workflow skills..."

**Challenge:** The number 15 is arbitrary and not grounded in evidence. The real test is **token cost per session**, not count. A 5KB skill used once per session is cheaper than a 500-byte skill loaded 100 times.

**Evidence from code:**
- `packages/cli/library/skills/` has 33 canonical skills at ~16KB total
- The Fortnite library has 37 skill directories at ~22KB total
- The speckit chain (7 skills: specify/plan/tasks/implement/checklist/analyze/clarify) alone is ~50KB

**Revised recommendation:** Use a **token-rent gate**: each skill must earn its place by being referenced in the default context OR being explicitly invoked. Don't set a count limit — set a byte-budget limit. Recommend **≤50KB total canonical library** (agents + skills + hooks combined).

### A.2.3 CHALLENGE: "Keep the Adapter Layer" — Coupling Risk Underestimated

**Prior claim:** "Preserve the adapter layer. It's the one piece of LazyAI that maps directly to vibe-lab's canonical → adapter architecture."

**Challenge:** The adapter layer has **41 Fortnite references** deeply embedded in `opencode.go`:  `fortnite/AGENTS.md` reading, `fortniteStartup` selection key, `fortniteScripts`, `fortniteWorkflows`, `fortniteAgents`, `fortniteSkills`. The adapter is not Fortnite-free — it's Fortnite-aware and Fortnite-coupled.

**Evidence:** `opencode.go` reads `fortnite/AGENTS.md` from library FS, selects Fortnite agents/skills/scripts/workflows by subdir, and sets `loop-driver` as default agent. Test file `adapter_test.go` (1,494 lines) heavily asserts Fortnite content exists.

**Revised recommendation:** The adapter **pattern** is correct, but the implementation needs **significant surgery** — not just "strip Fortnite-specific install paths." Specifically:
1. Remove `InstallToolContextFiles()` entirely (installs Fortnite context files)
2. Remove all `fortnite*` selection keys from `opencode.go:Install()`
3. Rewrite the `AGENTS.md` managed block to reference canonical agents, not Fortnite
4. Change default agent from `loop-driver` to a generic `orchestrator` or remove the default
5. **Estimated effort:** 2–3 days, not "low (remove ~50 files)"

### A.2.4 CHALLENGE: "Session CRUD Only" — Missing the Handoff Protocol

**Prior claim:** "Keep session CRUD only. Drop dispatch tracking, parallel tasks, barriers, locks, and messaging."

**Challenge:** Correct in stripping orchestration, but **incomplete**. A single-agent session that only does start/get/end is useless without a **handoff protocol** — the thing that makes session data durable across context switches.

**Evidence:** `session.go` currently has no `WriteHandoff()` method. The vibe-lab audit explicitly calls for handoff files with goal, status, files changed, commands run, validation results, blockers, and next steps. The 269-line session.go has zero handoff logic.

**Revised recommendation:** Session CRUD alone is ~100 lines of value. Add a `WriteHandoff(path string) error` method (~50 lines) that serializes the session state to a markdown handoff file. This makes the session actually useful rather than just a database record.

### A.2.5 NEW FINDING: The CLI Has 54 Command Files — Many Are Orchestration

The `packages/cli/cmd/` directory has **54 Go command files**. Many are orchestration-era commands:
- `orchestration.go` (539 lines) — chain/team/workflow management
- `workflow.go` (376 lines) — workflow execution
- `task.go` (271 lines) — task queue operations
- `message.go` (216 lines) — inter-agent messaging
- `metrics.go` (244 lines) — orchestration metrics
- `server.go` (851 lines) — orchestration server/dashboard
- `eval.go` (92 lines) — eval harness

**Risk:** Removing the Fortnite library and orchestrator package leaves **orphaned CLI commands** that compile but do nothing useful. These need to be removed too.

**Recommendation:** Audit all 54 CLI commands. Remove any that depend on removed packages. Estimate: **8–12 commands** need removal (~1,800 lines).

### A.2.6 NEW FINDING: Schema Has 20+ Tables — Most Are Orchestration

`packages/cli/internal/runtime/schema.go` defines 412 lines of SQL with **20+ tables**:
- **Keep (5):** `schema_migrations`, `sessions`, `ledgers`, `decisions`, `artifacts`
- **Drop (15+):** `dispatches`, `parallel_tasks`, `task_queue`, `task_messages`, `barriers`, `locks`, `messages`, `workflows`, `workflow_steps`, `teams`, `model_calls`, `quality_metrics`, `token_log`, `workflow_invocations`, `workflow_phases`, `workflow_agents`, `agent_configs`

**Risk:** A single `schema.go` file with 412 lines of mixed keep/drop tables. The migration system (`migrate.go`) needs to handle the downgrade — dropping unused tables without data loss for the kept ones.

**Recommendation:** Create a **clean V2 schema** (~80 lines) with only the 5 kept tables. Write a one-time migration script that exports session data, drops the DB, recreates with V2 schema, and re-imports sessions. ~2 hours of work.

### A.2.7 CHALLENGE: "Archive the Orchestrator Package" — Not Just Moving a Directory

**Prior claim:** "Archive the orchestrator package. It may have value for a future multi-tenant SaaS product."

**Challenge:** The orchestrator package (`packages/orchestrator/`) is a **separate Go module** with its own `go.mod`, `go.sum`, `.goreleaser.yaml`. It's not just "move dir" — it needs:
1. Remove `replace` directive from root `go.work` (if present)
2. Remove any workspace references
3. Move to `archive/` or a separate repo
4. Update any `import` paths that cross the boundary

**Evidence:** The internal orchestrator (`packages/cli/internal/orchestrator/`, 1,261 lines) is a separate concern — it provides catalog DB integration for the CLI. This also needs removal.

**Revised recommendation:** "Archive" is correct, but effort is **medium (half-day)**, not "low (move dir)."

---

## A.3 Revised Priority Matrix

| Rank | Action | Effort | Impact | Risk |
|------|--------|--------|--------|------|
| **P0** | Strip Fortnite library (146 files, 28,608 lines) | High (3 days) | Critical | Low — no Go code depends on Fortnite at runtime |
| **P0** | Remove orchestrator packages (19,436 lines) | Medium (half-day) | High | Medium — need to clean go.work and import paths |
| **P0** | Remove runtime orchestration (2,964 lines) | Medium (2 days) | High | Medium — CLI commands depend on these |
| **P0** | Remove orphaned CLI commands (~1,800 lines) | Low (1 day) | High | Low — breaking change, document in migration guide |
| **P1** | Surgery on adapter layer (Fortnite removal) | High (3 days) | Critical | High — tests will break, need rewrite |
| **P1** | Simplify schema to 5 tables | Low (2 hours) | Medium | Low — fresh start, no data migration needed |
| **P1** | Add handoff protocol to session | Low (1 day) | Medium | Low — new functionality, no breakage |
| **P2** | Shrink canonical library to ≤50KB | Medium (2 days) | Medium | Low — editorial work, no code changes |
| **P2** | Rewrite README | Low (1 hour) | Medium | Low — documentation only |
| **P3** | Remove Copilot adapter | Low (2 hours) | Low | Low — 520 lines, deferred target |

---

## A.4 What the Existing Report Gets Right

Despite the challenges above, the prior report's **core thesis is correct and well-supported**:

1. **Fortnite is the wrong foundation** — 28,608 lines of OpenCode-specific multi-agent orchestration that contradicts vibe-lab's "no heavy frameworks" principle.
2. **The adapter pattern is the innovation** — canonical → compile is the right architecture, just applied to the wrong surface.
3. **Session/ledger/DB core is reusable** — 1,331 lines of clean Go that map to vibe-lab's handoff protocol.
4. **The agent/skill surface is bloated** — 55 agents and 45 skills is a token bomb.
5. **The orchestrator package is speculative over-engineering** — 18,175 lines of hexagonal architecture for a personal lab.

---

## A.5 Execution Order — Recommended Phases

### Phase 1: Excision (Week 1)
1. Remove `packages/cli/library/fortnite/` entirely (146 files)
2. Remove `packages/orchestrator/` entirely (99 Go files)
3. Remove `packages/cli/internal/orchestrator/` (3 Go files)
4. Remove `packages/cli/internal/runtime/workflow/` (6 files)
5. Remove `packages/cli/internal/runtime/taskqueue/` (5 files)
6. Remove `packages/cli/internal/runtime/dispatch/` (3 files)
7. Remove session orchestration files: `dispatch.go`, `parallel.go`, `barrier.go`, `lock.go`, `message.go`
8. Remove orphaned CLI commands: `orchestration.go`, `workflow.go`, `task.go`, `message.go`, `metrics.go`, `server.go`, `eval.go`
9. **Verify:** `go build ./...` compiles. `go test ./...` passes (with failures in removed packages).

### Phase 2: Adapter Surgery (Week 2)
1. Remove all `fortnite*` references from `opencode.go:Install()`
2. Remove `InstallToolContextFiles()` from `shared.go`
3. Change default agent from `loop-driver` to generic or remove default
4. Update `adapter_test.go` to remove Fortnite assertions (~1,494 lines of tests heavily reference Fortnite)
5. Simplify schema to 5 tables
6. **Verify:** `go test ./internal/adapter/...` passes. `lazyai-cli init --target /tmp/test` produces clean output.

### Phase 3: Core Enhancement (Week 2–3)
1. Add `WriteHandoff()` to session package
2. Add PreToolUse safety hook (~60 lines)
3. Add instruction/data boundary blocks to agent/skill templates
4. **Verify:** `lazyai-cli session start "test goal" --handoff` writes a handoff file.

### Phase 4: Library Curation (Week 3)
1. Audit all 8 canonical agents against vibe-lab triggers
2. Audit all 33 canonical skills against token-rent gate
3. Remove or consolidate to ≤50KB total
4. Rewrite README
5. **Verify:** `du -sh packages/cli/library/agents/ packages/cli/library/skills/` ≤ 50KB.

---

## A.6 Final Verdict

The prior report is **directionally correct** but underestimates the adapter surgery effort and overestimates the "just archive" simplicity. The real work is in Phase 2 (adapter surgery) — that's where the Fortnite coupling lives in Go code, not just markdown files.

**Net code reduction:** ~49,880 lines removed, ~14,388 lines retained.
**Effort estimate:** 2–3 weeks for a single developer.
**Risk level:** Medium — breaking change for existing users, but clean architecture post-strip.

---

*Independent adversarial challenge complete. No implementation performed. Report only.*
