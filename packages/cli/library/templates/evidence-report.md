<template>
  <name>Evidence Report — Human Review</name>
  <output>specs/evidence-reports/YYYY-MM-DD-topic.md</output>
  <input>trace/failure + eval case + harness change + holdout review</input>
  <phase>Improvement — fifth and final step of the harness improvement loop</phase>
</template>

# Evidence Report: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**Human reviewer:** [name]
**Prepared by:** [name or agent]
**Status:** Draft | Under Review | Approved | Rejected | Promoted

> **Purpose.** Present the complete improvement loop to a human reviewer with all evidence collected. The human makes the final call: promote the asset update, reject it, or request changes. This is the only step in the loop that requires a human decision — everything before is automatable.

---

## 1. Loop Summary

| Step | Artifact | Status |
|---|---|---|
| 1. Trace / Failure | [`trace-failure.md`](link) | ✅ Captured |
| 2. Eval Case | [`eval-promotion-checklist.md`](link) | ✅ Defined |
| 3. Harness Change | [`harness-change-report.md`](link) | ✅ Applied |
| 4. Holdout Review | [`holdout-review.md`](link) | ✅ PASS |
| 5. Human Review | this file | ⬜ Pending |

**One-line summary:** [what failure was observed, what change was made, and why it fixes the root cause]

---

## 2. Evidence Package

### 2.1 Trace / Failure

| Field | Value |
|---|---|
| Category | [context / tooling / workflow / quality / adapter] |
| Tags | [comma-separated] |
| Severity | 🔴 / 🟡 / 🟢 / ⚪ |
| Frequency | [first / intermittent / recurring / continuous] |
| Raw evidence | [link to log, output, or screenshot] |

**Observation excerpt:**
```
[exact error, log line, or behavior — quoted from trace-failure.md]
```

### 2.2 Eval Case

| Field | Value |
|---|---|
| Trigger | Given [precondition], when [action], then [expected outcome] |
| PASS condition | [observable condition] |
| FAIL condition | [observable condition] |
| Tags | [comma-separated] |

### 2.3 Harness Change

| Field | Value |
|---|---|
| File(s) changed | [paths] |
| Change type | Template / Checklist / Standard / Prompt / Config / Other |
| Diff summary | [N lines added, N lines removed] |
| Hypothesis | If we [change], then [effect] |

**Change excerpt:**
```diff
[relevant diff — or link to full diff]
```

### 2.4 Holdout Review

| Field | Value |
|---|---|
| Reviewer | [name or agent — different from author] |
| Verdict | PASS / FAIL / Needs Revision |
| Blocking issues | [N — or "None"] |
| Regression check | [all existing eval cases still PASS / regressions found] |

**Reviewer comment:**
```
[reviewer's summary comment from holdout-review.md]
```

---

## 3. Human Review

### 3.1 Review Questions

The human reviewer should consider:

1. **Is the trace/failure real?** Is this a genuine harness gap, or an environment/one-off issue?
2. **Is the eval case well-defined?** Can a future agent or tool unambiguously determine PASS vs FAIL?
3. **Is the change minimal?** Does it fix exactly the root cause, or does it include speculative hardening?
4. **Is the change safe?** Does it risk regressing existing harness behavior?
5. **Is the change convention-fit?** Does it follow the existing template patterns and terminology?
6. **Is the change worth making?** Does the severity and frequency of the failure justify the change?

### 3.2 Human Decision

- [ ] **Approve** — promote the asset update. The change is correct, minimal, and safe.
- [ ] **Reject** — the change is not warranted. Rationale: [one sentence].
- [ ] **Request changes** — [specific changes needed before approval].

### 3.3 Human Notes

```
[free-form notes from the human reviewer]
```

---

## 4. Asset Update

If approved, record the asset update here.

| Asset | Action | New SHA / Version |
|---|---|---|
| [template file path] | create / modify / delete | [commit SHA or version] |
| [other affected file] | create / modify / delete | [commit SHA or version] |

**Promotion commit:** `[SHA]`

---

## 5. Closure

```
Status: APPROVED / REJECTED / CHANGES_REQUESTED
Human reviewer: [name]
Date: YYYY-MM-DD
Promotion commit: [SHA — if approved]
```

---

## 6. Post-Promotion (filled after promotion)

- [ ] Eval case status updated to `PROMOTED` in `eval-promotion-checklist.md`.
- [ ] Harness change status updated to `PROMOTED` in `harness-change-report.md`.
- [ ] Trace/failure status updated to `Resolved` in `trace-failure.md`.
- [ ] If the change introduces a new pattern: ADR created and linked.
- [ ] If the change affects user-visible behavior: CHANGELOG entry added.
