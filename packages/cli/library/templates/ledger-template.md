# Ledger — [repo-name]

> **Purpose.** Append-only durable record of every workflow that touched this repo. The ledger beats amnesia (Harness Rule 4 — Memory & State): future sessions read forward from the latest entry instead of re-deriving state.
>
> **Location:** `.specify/memory/repos/<repo-name>/ledger.md`.
> **Append-only.** Never delete entries. Corrections are appended as new entries that reference the prior date.

---

## How to append

Every workflow MUST add one entry on completion. Use the following block, copy-paste verbatim, fill the fields, and append at the **top** of the entries section (newest first).

```markdown
## YYYY-MM-DD — <type>(<scope>): <one-line summary>

**Who:** [author + reviewing agent]
**Workflow:** [speckit-implement / bugfix / spike / poc / housekeeping / refactor / review]
**Plan / spec:** [link to specs/###-slug/plan.md or spec.md]
**Status:** completed / partial / blocked / abandoned
**Verified:** [test command + result, OR human approval reference]

### Decisions
- [decision 1] — [why]
- [decision 2] — [why]

### Out-of-scope deferred
- [item] — [reason; pointer to follow-up if any]

### Follow-ups
- [ ] [open task] — [pointer]
- [ ] [open task] — [pointer]

### Lessons / standards
- [generalizable observation] — [proposed standards update if any]
```

---

## Entries (newest first)

<!-- New entries are appended above this line. -->

## YYYY-MM-DD — example(scope): seed entry, replace on first real append

**Who:** [name + agent]
**Workflow:** [workflow]
**Plan / spec:** [link]
**Status:** completed
**Verified:** [evidence]

### Decisions
- [decision] — [why]

### Out-of-scope deferred
- [item] — [reason]

### Follow-ups
- [ ] [pointer]

### Lessons / standards
- [observation]

---

## Index

A summary table of recent entries — regenerate with the `update-memory` skill on each append.

| Date | Workflow | Summary | Status | Plan link |
|---|---|---|---|---|
| YYYY-MM-DD | [workflow] | [summary] | completed | [link] |

> The index is a convenience for humans; the entries above are authoritative. If the index drifts from the entries, the entries win.
