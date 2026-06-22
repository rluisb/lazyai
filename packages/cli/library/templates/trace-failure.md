<template>
  <name>Trace / Failure Capture</name>
  <output>specs/traces/YYYY-MM-DD-topic.md</output>
  <input>observed failure, unexpected behavior, or notable trace event</input>
  <phase>Improvement — first step of the harness improvement loop</phase>
</template>

# Trace / Failure: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**Captured by:** [agent or human]
**Source:** [run / test / review / production / development]
**Status:** Open | Eval Case Created | Resolved | Closed

> **Purpose.** Capture a trace event or failure that may warrant a harness improvement. This is the raw observation — no analysis, no fix. The improvement loop starts here.

---

## 1. Observation

What happened. Be precise — include the exact command, output, or behavior.

```
[paste exact command, error, log lines, or observed behavior]
```

---

## 2. Context

What was being done when the event occurred.

| Aspect | Value |
|---|---|
| Task / workflow | [e.g., `rpi`, `bugfix`, `spec-authoring`] |
| Agent / tool | [e.g., `implementer`, `planner`, `human`] |
| Repo / branch | [repo, branch, commit SHA] |
| Harness version | [e.g., `v2.3` or commit of `constitution.md`] |
| Input artifact | [link to spec, plan, task harness, or prompt] |
| Output artifact | [link to output, log, or trace] |

---

## 3. Category

Classify the trace/failure using the trace taxonomy. Select one primary category; add secondary tags as needed.

| Category | Description | Select |
|---|---|---|
| **Context** | Missing, stale, or misleading context passed to an agent | ☐ |
| **Tooling** | Tool failure, missing capability, or incorrect tool use | ☐ |
| **Workflow** | Phase skipped, wrong order, or missing gate | ☐ |
| **Quality** | Output failed correctness, completeness, or convention checks | ☐ |
| **Adapter** | Target-specific output malformed or incompatible | ☐ |

**Primary category:** [one of the above]

**Tags:** [comma-separated, e.g. `context:missing-spec`, `tooling:lsp-failure`, `workflow:gate-skipped`, `quality:test-gap`, `adapter:opencode-format`]

---

## 4. Evidence

Supporting material. Attach or link to the raw evidence.

- [ ] Log file or command output: [path or artifact link]
- [ ] Screenshot or recording: [path or link]
- [ ] Related trace IDs: [list]
- [ ] Witness (human or agent who observed): [name]

---

## 5. Impact Assessment

| Dimension | Assessment |
|---|---|
| Severity | 🔴 Blocking / 🟡 Degrading / 🟢 Minor / ⚪ Informational |
| Frequency | First occurrence / Intermittent / Recurring / Continuous |
| Affected users | [agents, humans, or both] |
| Affected workflows | [list] |

---

## 6. Eval Case Mapping

Does this trace/failure map to an existing eval case, or should one be created?

- [ ] **Maps to existing eval case:** [eval case ID or path]
- [ ] **New eval case needed:** [brief description of the eval case]
- [ ] **Not an eval case** — this is a one-off or environment issue. Reason: [explain]

---

## 7. Next Action

- [ ] Create eval case from this trace (use `eval-promotion-checklist.md`)
- [ ] File as known issue (no immediate action)
- [ ] Close — no improvement warranted. Rationale: [one sentence]

---

## 8. Closure

```
Resolved: YYYY-MM-DD
Resolution: Eval Case Created / Filed / Closed
Eval case ref: [path or ID if created]
Closed by: [name or agent]
```
