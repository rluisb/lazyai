# Standard: Context Loading

**Category:** Process
**Scope:** project
**Date:** 2026-05-01
**Owner:** AI Setup Maintainers
**Status:** Active
**Constitution article(s):** III, IV, V

> **Purpose.** Keep AI and human review focused by loading the smallest useful set of instructions, contracts, and code evidence before making changes.

---

## Scope Cascade

This starter standard is written at project scope and applies to research, planning, implementation, review, and handoff work until a more specific context standard replaces it.

```
global standards
        ↓
workspace standards
        ↓
project standards
        ↓
task-specific context
```

**Where this standard lives:** project

**Overrides:** none — this is starter guidance for new projects.

---

## Rule

Before changing files, workers MUST load the task contract, relevant plan or spec section, local conventions, and directly affected implementation or test files; they MUST NOT bulk-load unrelated documentation.

**Trigger:** Starting a task, switching subtasks, changing a new module, or resolving a blocker that may alter scope.

---

## Rationale

Focused context reduces accidental scope creep and keeps decisions traceable to accepted contracts rather than stale or unrelated documents.

- **Article support:** Article III makes accepted specs and standards authoritative; Article IV prevents speculative exploration; Article V keeps the working set simple.
- **Origin:** Wave 1 starter standards for concrete pattern review.

---

## Examples

**Compliant:**
```
Read the task file, the matching acceptance criteria, the relevant plan section,
the local template, and the nearest validator test before editing.
```

**Non-compliant:**
```
Read every spec, all ADRs, and unrelated source packages before deciding what to edit.
```

**Why the non-compliant case fails:** It increases noise and makes it harder to prove that implementation followed the approved task scope.

---

## Enforcement

| Mechanism | Where | When |
|---|---|---|
| Pre-flight summary | Task response, handoff, or PR description | Before editing |
| Review against touch map | Git diff and task file | Gate 4 Pattern Consistency |
| Handoff notes | `specs/memory/handoffs/` when needed | At session boundary |

---

## Exceptions

Exceptions require a short note explaining why broader context is necessary and how the extra files affect the current decision.

- Cross-repo migration — broader context is allowed when the plan identifies multiple affected repos or contracts.

---

## Workspace Awareness

| Repo | Override | Reason |
|---|---|---|
| All repos | none | Starter guidance applies until repo-specific context-loading standards exist. |

---

## Related

- **Supersedes:** none
- **Related ADR(s):** none
- **Related standards:** orchestration-patterns, agent-security

---

## Memory

- [ ] Standard published at the declared scope.
- [ ] Recurring context-loading exceptions are captured as repo-specific guidance.
- [ ] If a missed context source causes rework, update the relevant task checklist or standard.
