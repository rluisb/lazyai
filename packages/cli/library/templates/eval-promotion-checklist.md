<template>
  <name>Eval Case Promotion Checklist</name>
  <output>specs/eval-cases/YYYY-MM-DD-topic.md</output>
  <input>trace/failure reference + tagged eval case definition</input>
  <phase>Improvement — second step of the harness improvement loop</phase>
</template>

# Eval Case: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**Author:** [name or agent]
**Trace/Failure ref:** [`trace-failure.md` path or ID]
**Status:** Draft | Ready for Holdout | Promoted | Retired

> **Purpose.** Define a tagged eval case from a trace or failure. An eval case is a repeatable check: given a specific input, the harness should produce a specific outcome. Eval cases are inspectable file assets — they do not imply a judge, scoring, or runtime engine.

---

## 1. Trace / Failure Reference

| Field | Value |
|---|---|
| Trace ID | [link to `trace-failure.md`] |
| Category | [context / tooling / workflow / quality / adapter] |
| Tags | [comma-separated, e.g. `context:missing-spec`, `quality:test-gap`] |
| Severity | 🔴 Blocking / 🟡 Degrading / 🟢 Minor / ⚪ Informational |
| Frequency | First / Intermittent / Recurring / Continuous |

---

## 2. Eval Case Definition

### 2.1 Trigger Condition

What input or scenario triggers this eval case.

```
Given: [precondition — e.g., "agent receives a spec with no acceptance criteria"]
When:  [action — e.g., "agent generates a task harness"]
Then:  [expected outcome — e.g., "harness includes a placeholder AC section"]
```

### 2.2 Pass / Fail Criteria

- **PASS:** [observable condition that means the harness handled this case correctly]
- **FAIL:** [observable condition that means the harness did not handle this case]

### 2.3 Example

```
Input:  [concrete example input]
Expected output: [what the harness should produce]
Actual output (before fix): [what the harness produced when the trace was captured]
```

---

## 3. Tagging

| Tag | Value | Purpose |
|---|---|---|
| `category` | [context/tooling/workflow/quality/adapter] | Primary taxonomy category |
| `severity` | [blocking/degrading/minor/info] | Impact level |
| `frequency` | [first/intermittent/recurring/continuous] | Occurrence pattern |
| `workflow` | [e.g. rpi, bugfix, spec-authoring] | Affected workflow |
| `agent` | [e.g. implementer, planner] | Affected agent role |
| `status` | [draft/ready/promoted/retired] | Lifecycle |

Additional tags (free-form):
- `[key]:[value]`
- `[key]:[value]`

---

## 4. Harness Change Hypothesis

What harness change would prevent this failure from recurring? (Filled before creating `harness-change-report.md`.)

> **If** we [proposed harness change],
> **then** this eval case would PASS.

---

## 5. Promotion Criteria

An eval case is ready for promotion when ALL of the following are true:

- [ ] **Trace is confirmed:** The trace/failure is real and reproducible (not a one-off environment glitch).
- [ ] **Category is correct:** The primary taxonomy category accurately describes the failure class.
- [ ] **Tags are complete:** At least `category`, `severity`, `frequency`, and `workflow` tags are set.
- [ ] **Pass/Fail criteria are observable:** A human (or future tool) can unambiguously determine PASS vs FAIL from the criteria alone.
- [ ] **Example is concrete:** The example input and expected output are specific enough to reproduce.
- [ ] **Harness change hypothesis is stated:** Even if the change hasn't been designed yet, the direction is clear.
- [ ] **No runtime dependency:** This eval case does not require a judge, scoring engine, or orchestration runtime to evaluate.

---

## 6. Holdout Review Reference

| Field | Value |
|---|---|
| Holdout review | [link to `holdout-review.md` — filled after review] |
| Holdout verdict | PASS / FAIL / Needs Revision |
| Promoted at | [date — filled after promotion] |

---

## 7. Verdict

```
Status: DRAFT / READY_FOR_HOLDOUT / PROMOTED / RETIRED
Eval case ID: [EC-### or slug]
Promoted by: [name or agent]
Date: YYYY-MM-DD
```
