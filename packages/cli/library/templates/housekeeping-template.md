# Housekeeping: [###-housekeeping-slug]

**ID:** ###
**Date:** YYYY-MM-DD
**Status:** Draft | In Progress | Complete
**Owner:** [name]
**Trigger:** scheduled / vulnerability / dependency drift / tech-debt review / release-prep

> **Purpose.** Structured maintenance: dependency updates, tech-debt resolution, cleanup. Housekeeping is **not** new feature work — it is preserving the project's ability to ship features tomorrow. Each item is small, isolated, and reversible.

---

## 1. Scope

What this housekeeping cycle covers. **One scope per housekeeping run.**

**Scope:** [dependencies / tech-debt / dead-code / docs-drift / config-cleanup / lint-debt / test-flakiness]

**Repos affected:** [list — usually one]

**Time-box:** [N hours / days]

---

## 2. Inventory

A complete list of items in scope. Score each by risk and value before working on any of them.

| Item | Type | Risk if ignored | Value of fixing | Effort | Priority |
|---|---|---|---|---|---|
| [item] | dep / debt / dead-code / doc / config | L \| M \| H | L \| M \| H | XS / S / M / L | P0 / P1 / P2 |

**Prioritization rule:** P0 = high risk OR security; P1 = high value, low effort; P2 = nice to have. Anti-Speculation (Article IV) — P3+ items are dropped, not stored.

---

## 3. Dependency Updates *(if scope = dependencies)*

| Package | Current | Target | Type | Breaking? | Notes |
|---|---|---|---|---|---|
| [name] | x.y.z | a.b.c | major / minor / patch / security | YES / NO | [migration notes] |

**Migration steps (per breaking dep):**
- [ ] Read CHANGELOG between versions.
- [ ] Identify call sites needing changes.
- [ ] Update + run test suite.
- [ ] Verify Gate 1-3 of the 5-gate ladder.

---

## 4. Tech-Debt Items *(if scope = tech-debt)*

Each item names a specific debt with a cited location.

| Item | Location | Why it is debt | Standard / ADR violated | Fix |
|---|---|---|---|---|
| [item] | [file:line or module] | [reason] | [pointer or "none — propose new"] | [planned change] |

> Article VI — Anti-Overengineering applies even to debt cleanup. Do not refactor "while we're here." Each row is its own commit.

---

## 5. Dead Code & Cleanup *(if scope = dead-code or config-cleanup)*

| Path | Evidence it is dead | Action |
|---|---|---|
| [path] | [grep result, codegraph trace, git-blame age] | DELETE / archive |

> Evidence MUST be objective (codegraph callers = 0, no imports, no test references). "I think it's dead" is not evidence.

---

## 6. Verification Strategy

Each housekeeping item is verified independently before moving to the next.

- [ ] Per-item tests run after the change.
- [ ] Per-item commit (atomic; reversible).
- [ ] After all items: full suite + integration tests.
- [ ] After all items: deploy to staging if applicable.

---

## 7. Quality Gate Checklist (5-Gate Ladder)

- [ ] **Gate 1 — Static Integrity:** lint, format, type-check on every changed file.
- [ ] **Gate 2 — Contract Compliance:** no API drift; deprecated callsites updated together.
- [ ] **Gate 3 — Behavioral Validation:** existing tests still pass; flaky-test debt acknowledged.
- [ ] **Gate 4 — Pattern Consistency:** no novel patterns introduced.
- [ ] **Gate 5 — Observability Readiness:** logging / metrics still functional after refactors.

---

## 8. Anti-Slope — Standards Updates

Did this housekeeping reveal a missing or outdated standard?

- [ ] **Pattern emerged:** [description] — proposed addition to `specs/standards/[file].md`.
- [ ] **Pattern rejected:** [description] — proposed addition to `specs/standards/[file].md` as "do not".
- [ ] **Standard outdated:** [pointer] — proposed amendment.

> Each fixed instance of debt should leave a trace so the same debt does not return.

---

## 9. Risks & Rollback

| Item | Risk | Rollback plan |
|---|---|---|
| [item] | [what could break] | [how to revert: `git revert SHA` / re-pin / restore file] |

---

## 10. Verdict

```
COMPLETE / PARTIAL / BLOCKED
Items closed:    N / M
Tests:           PASS / FAIL
Standards added: N
Date:            YYYY-MM-DD
Notes:           [one line]
```

---

## 11. Memory Update

- [ ] Append summary to `.specify/memory/repos/<repo>/ledger.md` (date, items closed, notes).
- [ ] Update `.specify/memory/repos/<repo>/last-known-state.md` if this changed branch / dirty files.
- [ ] Open follow-up tasks for items deferred (P2 → next cycle).
