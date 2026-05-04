# Checklist: [###-feature-slug]

**Feature ID:** ###
**Spec:** [./spec.md](./spec.md)
**Plan:** [./plan.md](./plan.md)
**Tasks:** [./tasks.md](./tasks.md)
**Date:** YYYY-MM-DD
**Status:** Draft | Validating | Passing | Failed
**Validator:** [name or agent]

> **Purpose.** This is the **objective evidence layer** for the feature. Every requirement in `spec.md`, every task in `tasks.md`, and every gate in the constitution gets a checkbox. Unchecked items at merge time block the merge. Checklists are "unit tests for English" — they catch ambiguity in the spec before it becomes ambiguity in the code.

---

## Spec Coverage

Every functional requirement from `spec.md` MUST have at least one passing test.

| FR ID | Requirement (short) | Test path | Pass? |
|---|---|---|---|
| FR-001 | [restate briefly] | [test path] | ☐ |
| FR-002 | [restate briefly] | [test path] | ☐ |

**Acceptance criteria coverage**
- [ ] Every P1 acceptance criterion has a passing test.
- [ ] Every P2 acceptance criterion (if in scope this phase) has a passing test.
- [ ] Every edge case (EC-*) has a passing test.

---

## Constitutional Conformance

Every relevant article gets a verdict. **Non-negotiable** articles (II, VI) MUST be PASS.

- [ ] **Article I — Library-First** — uses existing libraries; no parallel re-implementations.
- [ ] **Article II — Test-First (NON-NEGOTIABLE)** — git history shows test commit before code commit per task.
- [ ] **Article III — Docs as Source of Truth** — affected docs updated.
- [ ] **Article IV — Anti-Speculation** — no speculative code; out-of-scope items deferred not added.
- [ ] **Article V — Simplicity** — Complexity Tracking ledger reviewed; no unjustified entries.
- [ ] **Article VI — Anti-Overengineering (NON-NEGOTIABLE)** — no extraction below 3 instances; functions ≤ 30 lines; files ≤ 300 lines.

---

## 5-Gate Ladder

The ladder runs end-to-end on the merge candidate. A gate is **PASS** only with cited evidence.

### Gate 1 — Static Integrity
- [ ] Lint passes (no new warnings). Evidence: [command + result].
- [ ] Type check passes. Evidence: [command + result].
- [ ] Format clean. Evidence: [command + result].
- [ ] No dead code introduced.

### Gate 2 — Contract Compliance
- [ ] Public API matches `plan.md` Internal Contracts.
- [ ] Every `output_schema` section in produced artifacts is present.
- [ ] No undeclared side effects (filesystem / network / env).
- [ ] YAGNI / DRY-after-3 / KISS audit complete (Article VI).

### Gate 3 — Behavioral Validation
- [ ] Unit tests pass. Evidence: [command + count].
- [ ] Integration tests pass. Evidence: [command + count].
- [ ] Coverage meets project threshold. Evidence: [report link].
- [ ] (Bugfix only) Regression test fails on pre-fix code, passes on fix code. Evidence: [commit SHA pair].

### Gate 4 — Pattern Consistency
- [ ] Code reuses existing helpers (no parallel implementations).
- [ ] Naming follows project conventions.
- [ ] Error-handling pattern consistent with the codebase.
- [ ] No novel pattern introduced without ADR.

### Gate 5 — Observability Readiness
- [ ] Logging covers happy path and failure modes.
- [ ] Metrics exposed for the change's success/failure rate (if applicable).
- [ ] Errors carry enough context to debug from production logs.
- [ ] Rollback plan documented in PR description for risky changes.

---

## Memory & Workspace

- [ ] `update-memory` skill executed for the active repo.
- [ ] `ledger.md` entry appended (date, who, what, plan link, verified).
- [ ] `last-known-state.md` updated for the active repo.
- [ ] No PoC code shipped to main (Anti-Slope, Rule 5).

---

## Documentation

- [ ] `KNOWLEDGE_MAP.md` updated (if new module/feature).
- [ ] ADR written (if architectural decision made).
- [ ] Standards updated (if new pattern introduced and stable across 3+ instances).
- [ ] User-visible README / CHANGELOG entries added (if user-visible change).

---

## Verdict

```
PASS / FAIL — [N gates failed, N articles failed]
Validator: [name or agent]
Date:      YYYY-MM-DD
Notes:     [one or two sentences]
```

> A FAIL verdict blocks merge. The author returns to the failing gate and either fixes the underlying issue or escalates to a human gate with justification.
