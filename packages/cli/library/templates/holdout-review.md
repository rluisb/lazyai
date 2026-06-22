<template>
  <name>Holdout Review</name>
  <output>specs/holdout-reviews/YYYY-MM-DD-topic.md</output>
  <input>harness change report + eval case + existing harness state</input>
  <phase>Improvement — fourth step of the harness improvement loop</phase>
</template>

# Holdout Review: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**Reviewer:** [name or agent — MUST be different from the change author]
**Harness change ref:** [`harness-change-report.md` path or ID]
**Eval case ref:** [`eval-promotion-checklist.md` path or ID]
**Status:** Pending | PASS | FAIL | Needs Revision

> **Purpose.** A holdout review is the fourth step in the harness improvement loop. Before a harness change is promoted, a reviewer who did NOT author the change verifies that (a) the change is correct and minimal, (b) the eval case passes with the change applied, and (c) the change does not regress existing behavior. The reviewer is the gate — no change reaches the harness without a passing holdout review.

---

## 1. Change Under Review

| Field | Value |
|---|---|
| Change report | [link] |
| Eval case | [link] |
| Trace/failure | [link] |
| Files changed | [list of files and their diff summary] |
| Change type | Template / Checklist / Standard / Prompt / Config / Other |

---

## 2. Review Criteria

### 2.1 Correctness

- [ ] **Change matches hypothesis:** The change in `harness-change-report.md` §3 matches the hypothesis in §1.
- [ ] **Eval case passes:** With the change applied, the eval case trigger produces the expected PASS outcome.
- [ ] **Eval case fails without change:** Confirmed that the eval case would FAIL on the pre-change harness.
- [ ] **No side effects:** The change does not break other eval cases or harness behavior.

### 2.2 Minimality

- [ ] **Single cause:** The change addresses exactly one root cause (no bundled fixes).
- [ ] **No speculative hardening:** No defensive additions beyond the root cause.
- [ ] **No unrelated changes:** No formatting, renames, or refactors outside the scope.
- [ ] **Smallest possible diff:** The change could not be smaller while still fixing the root cause.

### 2.3 Reversibility

- [ ] **Clean revert:** The change can be reverted in one commit.
- [ ] **No collateral:** Reverting this change does not affect unrelated harness behavior.
- [ ] **Rollback plan documented:** `harness-change-report.md` §7 is complete.

### 2.4 Convention Fit

- [ ] **Follows template conventions:** Format, frontmatter, and section structure match existing templates.
- [ ] **Consistent terminology:** Uses the same terms as the rest of the harness (e.g., "trace/failure", "eval case", "holdout review").
- [ ] **No new patterns:** Does not introduce a novel template structure or convention without ADR.

---

## 3. Holdout Test

Run the eval case against the changed harness. Record the result.

| Test | Pre-change result | Post-change result | Verdict |
|---|---|---|---|
| Eval case [ID] | FAIL | PASS | ✅ |
| [Existing eval case 1] | PASS | PASS | ✅ |
| [Existing eval case 2] | PASS | PASS | ✅ |

**Regression check:** All existing eval cases still PASS after the change.

---

## 4. Findings

### 🔴 Blocking (must fix before promotion)

- [ ] [finding] — [file:line or section reference]

### 🟡 Concerns (should address)

- [ ] [finding] — [file:line or section reference]

### 🟢 Notes (no action required)

- [ ] [observation]

---

## 5. Verdict

```
HOLDOUT REVIEW: PASS / FAIL / NEEDS REVISION

Blocking issues: [N — or "None"]
Concerns:       [N — or "None"]
Notes:          [N — or "None"]

Reviewer: [name or agent]
Date:      YYYY-MM-DD
```

**PASS** — The change is correct, minimal, reversible, and convention-fit. Proceed to human review.

**FAIL** — The change has blocking issues. Return to `harness-change-report.md` for revision.

**NEEDS REVISION** — Minor issues that the author can fix without a full re-review. Author addresses findings in §4, then the reviewer confirms.

---

## 6. Next Step

- [ ] **Promote to human review** — change + holdout review forwarded for `evidence-report.md`.
- [ ] **Return for revision** — author addresses blocking findings.
- [ ] **Reject** — change is not warranted. Rationale: [one sentence].
