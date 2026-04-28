# Plan: Spec 022 — Workflow Library Alignment

**Status:** In Plan | **Date:** 2026-04-28 | **Spec:** [spec.md](./spec.md)

## Summary

Rebuild ai-setup's library (~45 files created/updated) to provide 13 speckit-aligned workflows, embed Harness Engineering + prompt techniques, and enforce anti-overengineering + TDD-first policies.

## Phase Ordering

Phases are ordered by dependency: foundation first (constitution, fragments), then templates (consumed by skills), then skills and agents (consume templates and fragments), then scaffold/compile integration (consumes all library content), then verification.

---

## Phase A: Constitution & Fragments (Foundation)

**Why first:** Constitution and fragments are read by every skill and agent. They define the policy layer.

### A1 — Rebuild constitution template

Replace 4 separate files (`constitution.template.md`, `constraints.template.md`, `quality-gates.template.md`, `uncertainty.template.md`) with 1 consolidated file.

| Task | File | Effort |
|---|---|---|
| A1.1 | Merge + rewrite `library/constitution/constitution.template.md` | M |
| A1.2 | Add Article VI: Anti-Overengineering (YAGNI, DRY, KISS, Clean Code, Unix) | M |
| A1.3 | Add TDD-first policy | S |
| A1.4 | Delete `constraints.template.md`, `quality-gates.template.md`, `uncertainty.template.md` | S |

**Output:** 1 file (was 4). Single constitution with: Core Principles, Technology Constraints, Anti-Overengineering, Quality Gates, TDD Policy, Governance.

### A2 — Rebuild quality gates fragment

| Task | File | Effort |
|---|---|---|
| A2.1 | Rewrite `library/fragments/quality-gates.xml` to 5-gate ladder | M |
| A2.2 | Add pre-commit, contract compliance, behavioral, pattern consistency, observability gates | M |
| A2.3 | Add YAGNI/DRY/KISS checks to Gate 2 (Contract Compliance) | S |

### A3 — Create new fragments

| Task | File | Effort |
|---|---|---|
| A3.1 | Create `library/fragments/harness-protocol.md` (Feed Forward, Contract, Feedback, Memory, Anti-Slope with concrete case studies) | L |
| A3.2 | Create `library/fragments/workspace-protocol.md` (multi-repo awareness, ledger updates, standards propagation) | M |

### A4 — Update existing fragments

| Task | File | Effort |
|---|---|---|
| A4.1 | Rebuild `library/fragments/rpi-workflow.md` with Harness phases (Feed Forward → Research → Plan → Implement → Feedback) | M |
| A4.2 | Update `library/fragments/reasoning-protocol.md` with CoT, ReAct, ToT, LLM-as-Judge technique references | S |

**Phase A total:** 9 tasks, ~M effort

---

## Phase B: Templates (Consumed by Skills & Agents)

### B1 — Speckit-aligned templates

| Task | File | Effort |
|---|---|---|
| B1.1 | Rewrite `library/templates/spec-template.md` — speckit format (User Scenarios P1-P3, FR-*, Success Criteria, Edge Cases, Assumptions) | M |
| B1.2 | Rewrite `library/templates/plan-template.md` — speckit format (Technical Context, Constitution Check, Project Structure, Complexity Tracking) | M |
| B1.3 | Rewrite `library/templates/tasks-template.md` — speckit format (Phases, [P] markers, [US*] labels, Dependencies) | M |
| B1.4 | Rewrite `library/templates/checklist-template.md` — speckit quality gate format | S |

### B2 — Task harness template (new)

| Task | File | Effort |
|---|---|---|
| B2.1 | Create `library/templates/task-harness-template.md` — 5-gate ladder, environment (tool versions), pattern references, testing strategy, permissions | M |

### B3 — New workflow templates

| Task | File | Effort |
|---|---|---|
| B3.1 | Create `library/templates/spike-template.md` | M |
| B3.2 | Create `library/templates/poc-template.md` | S |
| B3.3 | Create `library/templates/housekeeping-template.md` | S |
| B3.4 | Create `library/templates/ledger-template.md` | S |
| B3.5 | Create `library/templates/audit-template.md` | S |

### B4 — Update existing templates

| Task | File | Effort |
|---|---|---|
| B4.1 | Update `library/templates/adr.md` — add speckit constitution reference | S |
| B4.2 | Update `library/templates/bugfix-rca-template.md` — add Harness Contract section | S |
| B4.3 | Update `library/templates/standard.md` — workspace-aware format | S |
| B4.4 | Update `library/templates/code-review-template.md` — add simplicity audit section | S |

**Phase B total:** 13 tasks, ~M effort

---

## Phase C: Skills (Workflow Definitions)

### C1 — Speckit core skills (8 new/rebuilt)

| Task | File | Effort |
|---|---|---|
| C1.1 | Create `library/skills/speckit-constitution.md` — 9-section blueprint, establishes project principles | L |
| C1.2 | Create `library/skills/speckit-specify.md` — includes scout pre-flight for spec numbering, CoT for user stories, Feed Forward for blueprint | L |
| C1.3 | Create `library/skills/speckit-clarify.md` — sequential questioning, records Clarifications section | M |
| C1.4 | Rebuild `library/skills/speckit-plan.md` from `plan.md` — ReAct for tech stack, Constitution Check, research.md + data-model.md + quickstart.md generation | L |
| C1.5 | Create `library/skills/speckit-tasks.md` — Prompt Chaining (plan → tasks), per-user-story phases, [P] markers | L |
| C1.6 | Create `library/skills/speckit-analyze.md` — Self-Consistency for cross-artifact check, Contract verification | M |
| C1.7 | Create `library/skills/speckit-checklist.md` — quality gate checklists, "unit tests for English" | M |
| C1.8 | Rebuild `library/skills/speckit-implement.md` from `implement.md` — embedded TDD loop, 5-gate ladder per task, implements from task harness | L |

### C2 — ai-setup extension skills (8 new/rebuilt)

| Task | File | Effort |
|---|---|---|
| C2.1 | Update `library/skills/rpi.md` — orchestration chain (research → plan → implement), Harness phases, lighter than SDD | M |
| C2.2 | Create `library/skills/bugfix.md` — RCA → hypothesis → fix → verify, Contract (expected vs actual) | M |
| C2.3 | Create `library/skills/spike.md` — Tree-of-Thoughts, Self-Consistency, Generated Knowledge, graphify MCP | M |
| C2.4 | Create `library/skills/proof-of-concept.md` — Mini-RPI, Anti-Slope (throw away PoC), success criteria | M |
| C2.5 | Create `library/skills/housekeeping.md` — dependency updates, tech debt, structured output | M |
| C2.6 | Rebuild `library/skills/review.md` — LLM-as-Judge pattern, 5 lenses (Test Quality, Contract, Patterns, Perf/Security, Simplicity Audit), codegraph + qmd MCP | L |
| C2.7 | Update `library/skills/extract-standards.md` — workspace-aware, codegraph + qmd MCP | M |
| C2.8 | Create `library/skills/update-memory.md` — ledger append, state snapshot, obsidian + memory MCP | M |
| C2.9 | Create `library/skills/self-improve.md` — Reflexion, ART, graphify + obsidian MCP | M |
| C2.10 | Create `library/skills/process-audit.md` — Self-Consistency, structured audit report, qmd MCP | M |

### C3 — Keep existing skills (unchanged)

- `tdd-loop.md` — Keep (standalone TDD for non-speckit tasks)
- `parallel-execution.md` — Keep
- `anti-speculation.md` — Keep
- `memory-write.md` — Deprecate (replaced by `update-memory.md`), keep for backward compat

**Phase C total:** 20 tasks, ~M-L effort

---

## Phase D: Agents (Updated + New)

### D1 — New agent: implementor

| Task | File | Effort |
|---|---|---|
| D1.1 | Create `library/agents/implementor.md` — TDD-first, 5-gate ladder, task harness consumer, codegraph for pattern matching, anti-overengineering constraints | L |

### D2 — Update existing agents

| Task | File | Effort |
|---|---|---|
| D2.1 | Update `library/agents/scout.md` — add pre-flight spec numbering, speckit-aware research | M |
| D2.2 | Update `library/agents/planner.md` — speckit-aware plan/task generation, constitution check | M |
| D2.3 | Update `library/agents/builder.md` — workspace awareness (respects per-repo permissions, updates ledgers) | M |
| D2.4 | Update `library/agents/reviewer.md` — LLM-as-Judge synthesizer, 5 lenses (Test Quality first, Simplicity Audit), evidence-based, codegraph + qmd MCP | L |
| D2.5 | Update `library/agents/orchestrator.md` — Multi-Agent Orchestrator topology (decompose → workers → synthesize), dynamic worker assignment | L |
| D2.6 | Update `library/agents/red-team.md` — Dual-Agent Contract mode, adversarial spec/code verification | M |
| D2.7 | Update `library/agents/documenter.md` — minor updates for workspace awareness | S |

**Phase D total:** 8 tasks, ~M-L effort

---

## Phase E: Scaffolding & Compile Integration (ai-setup Code Changes)

### E1 — Scaffold spec-kit structure

| Task | File | Effort |
|---|---|---|
| E1.1 | Rewrite `src/scaffold/specs.ts` + `internal/scaffold/specs.go` — generate `specs/.gitkeep` + `.specify/` templates instead of flat categories | L |
| E1.2 | Add spec-kit structure detection (`.specify/` and `specs/###-slug/`) — never overwrite existing structures | M |
| E1.3 | Update `src/scaffold/repo-roots.ts` + `internal/scaffold/repos.go` — add `.specify/memory/repos/<name>/` ledger creation | M |

### E2 — Compile integration

| Task | File | Effort |
|---|---|---|
| E2.1 | Update adapter output mapping — ensure skills generate correct tool-native files for all 6+ tools | M |
| E2.2 | Add output contract validation — verify `output` and `produces_for` chains during compile | M |
| E2.3 | Add workspace root support — generate AI tool configs at workspace root when `--scope workspace` | L |

### E3 — Library structure reorg

| Task | File | Effort |
|---|---|---|
| E3.1 | Move/deprecate old files, ensure `ai-setup compile` picks up new skill names | M |
| E3.2 | Update `library/mcp/catalog.json` — ensure graphify, codegraph, qmd, obsidian are in supported MCP list | S |

**Phase E total:** 8 tasks, ~M-L effort

---

## Phase F: Testing & Verification

### F1 — New tests

| Task | Effort |
|---|---|
| F1.1 | Skill frontmatter validation tests — verify all skills declare techniques, output schema, consumed files, MCP tools | M |
| F1.2 | Template format validation — verify speckit templates match expected sections | M |
| F1.3 | Output contract validation — verify skill chains are consistent (output of A matches input of B) | M |
| F1.4 | Workspace-aware skill behavior — verify ledgers created, existing structures respected | M |
| F1.5 | Pre-flight spec numbering test — verify scout detects PR collisions | S |
| F1.6 | Anti-overengineering gate test — verify reviewer flags YAGNI/DRY violations | S |

### F2 — Regression

| Task | Effort |
|---|---|
| F2.1 | Run full test suite (pnpm + go) — ensure no regression in scaffold, compile, or adapter logic | S |

### F3 — Documentation

| Task | Effort |
|---|---|
| F3.1 | Update `README.md` with new workflow catalog | S |
| F3.2 | Create `docs/workflows.md` with usage guide | M |

**Phase F total:** 9 tasks, ~M effort

---

## Dependency Graph

```
Phase A (Constitution + Fragments)
    │
    ▼
Phase B (Templates) ────────────────────┐
    │                                    │
    ▼                                    ▼
Phase C (Skills) ◄──────────── Phase D (Agents)
    │                                    │
    └──────────┬─────────────────────────┘
               ▼
        Phase E (Scaffolding + Compile)
               │
               ▼
        Phase F (Testing + Verification)
```

**Key dependencies:**
- A must complete before B (templates reference constitution)
- B must complete before C + D (skills and agents consume templates)
- C + D can run in parallel (skills and agents are independent files)
- E depends on C + D (compile needs final skill/agent paths)
- F runs after all phases

---

## Effort Summary

| Phase | Tasks | Effort |
|---|---|---|
| A — Constitution & Fragments | 9 | M (3-4h) |
| B — Templates | 13 | M (3-4h) |
| C — Skills | 20 | M-L (5-7h) |
| D — Agents | 8 | M-L (3-5h) |
| E — Scaffolding & Compile | 8 | M-L (4-6h) |
| F — Testing & Verification | 9 | M (3-4h) |
| **Total** | **67** | **~21-30h** |

---

## File Manifest

### Created (27 new files)
```
library/skills/speckit-constitution.md
library/skills/speckit-specify.md
library/skills/speckit-clarify.md
library/skills/speckit-tasks.md
library/skills/speckit-analyze.md
library/skills/speckit-checklist.md
library/skills/bugfix.md
library/skills/spike.md
library/skills/proof-of-concept.md
library/skills/housekeeping.md
library/skills/update-memory.md
library/skills/self-improve.md
library/skills/process-audit.md
library/agents/implementor.md
library/templates/task-harness-template.md
library/templates/spike-template.md
library/templates/poc-template.md
library/templates/housekeeping-template.md
library/templates/ledger-template.md
library/templates/audit-template.md
library/fragments/harness-protocol.md
library/fragments/workspace-protocol.md
```

### Rebuilt (10 files — rewritten from scratch)
```
library/constitution/constitution.template.md  (was 4 files → 1)
library/fragments/quality-gates.xml
library/fragments/rpi-workflow.md
library/templates/spec-template.md
library/templates/plan-template.md
library/templates/tasks-template.md  (was task.md)
library/templates/checklist-template.md
library/skills/speckit-plan.md  (was plan.md)
library/skills/speckit-implement.md  (was implement.md)
library/skills/review.md
```

### Updated (16 files — modified in place)
```
library/fragments/reasoning-protocol.md
library/templates/adr.md
library/templates/bugfix-rca-template.md
library/templates/standard.md
library/templates/code-review-template.md
library/skills/rpi.md
library/skills/extract-standards.md
library/skills/tdd-loop.md
library/agents/scout.md
library/agents/planner.md
library/agents/builder.md
library/agents/reviewer.md
library/agents/orchestrator.md
library/agents/red-team.md
library/agents/documenter.md
library/mcp/catalog.json
```

### Deprecated (3 files — kept for backward compat)
```
library/skills/memory-write.md   (replaced by update-memory.md)
library/constitution/constraints.template.md   (merged into constitution)
library/constitution/quality-gates.template.md  (merged into constitution)
library/constitution/uncertainty.template.md    (merged into constitution)
```

### Code changes (TS + Go)
```
src/scaffold/specs.ts              (rewrite)
src/scaffold/repo-roots.ts         (update)
src/__tests__/ (new test files)    (create)
internal/scaffold/specs.go         (rewrite)
internal/scaffold/repos.go         (update)
```

---

## Acceptance Criteria (from spec)

1. ✅ All 13 workflows have corresponding skills
2. ✅ Speckit-aligned templates exist with matching format
3. ✅ Each skill declares techniques, output schema, consumed files, MCP tools
4. ✅ All skills follow 9-section blueprint with reasoning-model variant
5. ✅ Harness Engineering concepts at appropriate phases
6. ✅ Output contracts enable Prompt Chaining
7. ✅ All agents updated for speckit + workspace + LLM-as-Judge
8. ✅ Orchestrator implements Worker topology
9. ✅ RPI fragment rebuilt with Feed Forward → Research → Plan → Implement → Feedback
10. ✅ Human-in-the-Loop formalized
11. ✅ Workspace integration — ledgers, standards propagation
12. ✅ Existing spec-kit structures detected and respected
13. ✅ Tool-native output for all 6+ tools
14. ✅ MCP tools only — no external paid services
15. ✅ Existing tests pass — no regression
16. ✅ New tests cover frontmatter, templates, contracts, workspace behavior
