<template>
  <name>Harness Change Report</name>
  <output>specs/harness-changes/YYYY-MM-DD-topic.md</output>
  <input>trace/failure reference + eval case + proposed harness change</input>
  <phase>Improvement — third step of the harness improvement loop</phase>
</template>

# Harness Change: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**Author:** [name or agent]
**Trace/Failure ref:** [`trace-failure.md` path or ID]
**Eval case ref:** [`eval-promotion-checklist.md` path or ID]
**Status:** Draft | Applied | Verified | Promoted | Reverted

> **Purpose.** Document one targeted harness change. Every harness change is a hypothesis: "this change will prevent the trace/failure from recurring." The change is minimal, scoped, and reversible.

---

## 1. Hypothesis

One sentence: what change to the harness will prevent this class of failure from recurring.

> **If** we [change to harness],
> **then** [expected effect on trace/failure class].

---

## 2. Trace / Failure Link

| Field | Value |
|---|---|
| Trace ID | [link to `trace-failure.md`] |
| Category | [context / tooling / workflow / quality / adapter] |
| Tags | [comma-separated] |
| Root cause (from eval case) | [one sentence] |

---

## 3. Change Description

| Aspect | Detail |
|---|---|
| **What changed** | [file path + description of the change] |
| **Type** | Template / Checklist / Standard / Prompt / Config / Other |
| **Scope** | [single file / single section / single rule] |
| **Before** | [what the harness said or did before] |
| **After** | [what the harness says or does after] |

**Diff or change excerpt:**

```diff
[show the change — or paste the relevant before/after section]
```

---

## 4. Scope Guard (Anti-Speculation Check)

Every harness change MUST pass this guard. A change that fails any check is too broad.

- [ ] **Single cause:** This change addresses exactly one root cause from the eval case.
- [ ] **No speculative hardening:** No defensive additions "while we're here."
- [ ] **No unrelated cleanup:** No formatting, renames, or refactors outside the change.
- [ ] **Reversible:** The change can be reverted in one commit without collateral.
- [ ] **Measurable:** We can tell whether the change had the intended effect (fewer traces of this category).

**Scope verdict:** PASS / FAIL — [if FAIL, split into separate changes]

---

## 5. Verification

How was the change verified before proceeding to holdout review?

- [ ] **Template renders correctly:** [command or inspection result]
- [️] **Existing tests pass:** [command + result]
- [ ] **No new lint/format warnings:** [command + result]
- [ ] **Change is self-consistent:** no dangling references, no broken links
- [ ] **Holdout review prepared:** `holdout-review.md` created

---

## 6. Related Changes

| File | Action | Rationale |
|---|---|---|
| [path] | create / modify / delete | [why this file is part of the change] |

**Out of bounds (explicitly NOT changed):**
- [path or pattern] — [reason]

---

## 7. Rollback Plan

If this change causes regressions:

- **Revert command:** `git revert <SHA>`
- **Detection:** [how we would notice — e.g., "eval case X fails", "trace category Y increases"]
- **Fallback state:** [what the harness looked like before]

---

## 8. Verdict

```
Status: APPLIED / VERIFIED / PROMOTED / REVERTED
Applied at: [commit SHA]
Verified by: [name or agent]
Date: YYYY-MM-DD
```
