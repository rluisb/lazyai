# Trace Taxonomy: Evidence Classification for Human/Host-Tool Review

**Document type:** Template — inspectable taxonomy asset
**Purpose:** Classify agent-session evidence traces into categories and tags for structured review, failure analysis, and improvement-loop feed-in.
**Boundary:** This is an **evidence taxonomy for human and host-tool review** — not a trace daemon, not a runtime capture engine, not a scoring/judge runtime. Traces are produced by the host tool's own session recording; this taxonomy classifies what those traces contain.

---

## 1. Categories

Every trace evidence item belongs to exactly one category. Categories are mutually exclusive.

| Category | Definition | What it captures |
|---|---|---|
| `context` | What the agent knew before acting | Loaded files, environment snapshot, instructions, spec references, prior decisions, assumptions recorded |
| `tooling` | How the agent interacted with tools | Tool invocations, parameters, results, errors, timeouts, retries, fallback chains |
| `workflow` | The process the agent followed | Phase transitions, gate checks, handoffs, human-gate stops, lifecycle label changes |
| `quality` | What the agent produced and how it verified | Test results, lint/build output, evidence records, diff quality, acceptance-criterion satisfaction |
| `adapter` | How the agent adapted to project constraints | Pattern following, convention application, constraint honoring, speculative-code avoidance, escalation decisions |

### Category selection rules

- If the evidence item describes **what the agent read or knew** → `context`
- If the evidence item describes **a tool call or its outcome** → `tooling`
- If the evidence item describes **a process step, phase, or gate** → `workflow`
- If the evidence item describes **output verification or quality assessment** → `quality`
- If the evidence item describes **how the agent adjusted to constraints or conventions** → `adapter`

---

## 2. Tags

Tags are the granular classification within a category. An evidence item MUST have exactly one category and MAY have zero or more tags.

### 2.1 Context tags

| Tag | Meaning |
|---|---|
| `spec_loaded` | Agent read a spec, plan, or requirements document |
| `codebase_explored` | Agent searched, read, or explored the codebase |
| `env_snapshot` | Agent recorded or referenced environment state (OS, tool versions, env vars) |
| `constraints_known` | Agent acknowledged project constraints (boundaries, non-goals, out-of-bounds) |
| `assumptions_recorded` | Agent explicitly recorded an assumption or uncertainty |
| `prior_decision` | Agent referenced a prior decision, ADR, or memory entry |
| `instructions_loaded` | Agent read skill, rule, or instruction files |

### 2.2 Tooling tags

| Tag | Meaning |
|---|---|
| `tool_invoked` | Agent called a tool (read, search, edit, bash, etc.) |
| `tool_result` | Agent received and processed a tool result |
| `tool_error` | Tool returned an error or unexpected result |
| `tool_timeout` | Tool call exceeded its timeout |
| `tool_retry` | Agent retried a tool call after failure |
| `tool_fallback` | Agent switched to an alternative tool or approach |
| `tool_chain` | Agent used multiple tools in sequence for one logical operation |

### 2.3 Workflow tags

| Tag | Meaning |
|---|---|
| `phase_start` | Agent entered a new workflow phase |
| `phase_complete` | Agent completed a workflow phase |
| `gate_passed` | Agent passed a quality or process gate |
| `gate_failed` | Agent failed a quality or process gate |
| `human_gate` | Agent stopped for human approval (⛔ gate) |
| `handoff` | Agent produced or consumed a handoff document |
| `lifecycle_change` | Agent changed its lifecycle label |
| `escalation` | Agent escalated a decision outside its scope |

### 2.4 Quality tags

| Tag | Meaning |
|---|---|
| `test_passed` | A test passed |
| `test_failed` | A test failed |
| `lint_ok` | Lint/format check passed |
| `lint_error` | Lint/format check failed |
| `build_ok` | Build/type-check passed |
| `build_error` | Build/type-check failed |
| `evidence_recorded` | Agent recorded verification evidence |
| `ac_satisfied` | An acceptance criterion was satisfied |
| `ac_failed` | An acceptance criterion was not satisfied |
| `regression_detected` | A change introduced a regression |

### 2.5 Adapter tags

| Tag | Meaning |
|---|---|
| `pattern_followed` | Agent followed an existing code pattern or convention |
| `convention_applied` | Agent applied a project naming/style/testing convention |
| `constraint_honored` | Agent respected a stated constraint (out-of-bounds, non-goal) |
| `speculative_avoided` | Agent explicitly avoided speculative or out-of-scope work |
| `escalation` | Agent escalated a design decision outside task scope |
| `minimal_change` | Agent made the smallest change that satisfies the requirement |

---

## 3. Evidence Record Format

Each trace evidence item SHOULD be recorded as a structured entry:

```json
{
  "category": "context | tooling | workflow | quality | adapter",
  "tags": ["tag1", "tag2"],
  "timestamp": "YYYY-MM-DDTHH:MM:SSZ",
  "phase": "research | plan | implement | verify | cleanup",
  "summary": "One-line description of what happened",
  "evidence": "Path, command output, or reference to the supporting artifact",
  "outcome": "success | failure | partial | blocked"
}
```

### Minimal form (for inline use in handoffs or status reports)

```
[category:tag1,tag2] summary — evidence: <path or ref>
```

Example:
```
[context:spec_loaded,constraints_known] Loaded spec.md and noted out-of-bounds paths — evidence: spec.md:12-15
[tooling:tool_invoked,tool_result] Searched for existing pattern in src/utils/ — evidence: search result: 3 matches
[workflow:gate_passed] Gate 1 (static integrity) passed — evidence: go vet: OK
[quality:test_failed,regression_detected] Test TestParseConfig failed — evidence: go test ./internal/config: FAIL
[adapter:pattern_followed,minimal_change] Used existing Config.Load pattern instead of new constructor — evidence: src/config.go:42
```

---

## 4. Improvement Loop Integration

The trace taxonomy feeds the improvement loop as follows:

```
trace evidence (from session)
  → classify by category + tags (this taxonomy)
    → identify failure pattern (e.g., recurring test_failed + regression_detected)
      → create targeted eval case (one harness change)
        → run against holdout set
          → human review
            → promote asset update (skill, rubric, template, or invariant)
```

This taxonomy is the **classification layer** only. The scoring, holdout, and promotion machinery is defined separately in the eval rubric assets and the improvement-loop concept document.

### What this taxonomy is NOT

- **Not a trace daemon** — no background process, no automatic capture, no runtime overhead.
- **Not a scoring engine** — no judge LLM, no numeric scores, no pass/fail thresholds.
- **Not an orchestration layer** — no workflow engine, no subagent spawning, no state machine.
- **Not a replacement for host-tool session logs** — it classifies what the host tool already records.

---

## 5. Relationship to Other Assets

| Asset | Relationship |
|---|---|
| [Eval rubrics](../rubrics/README.md) | Rubrics define pass/fail criteria; this taxonomy classifies the evidence they evaluate |
| [Improvement-loop concept](../../../docs/concepts/trace-eval-improvement-loop.md) | Describes the full loop: trace → classify → eval → promote |
| [Task harness template](./task-harness-template.md) | Harness templates record trace evidence in their quality-gate sections |
| [Audit template](./audit-template.md) | Audits use classified traces to measure workflow adherence |
| [Handoff template](./handoff-template.md) | Handoffs carry minimal-form trace evidence for the next agent |
